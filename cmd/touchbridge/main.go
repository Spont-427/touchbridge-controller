package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Spont-427/touchbridge-controller/internal/discovery"
	"github.com/Spont-427/touchbridge-controller/internal/input"
	"github.com/Spont-427/touchbridge-controller/internal/web"
)

const appVersion = "0.5.2-alpha"

func main() {
	listen := flag.String("listen", "0.0.0.0:8765", "HTTP listen address")
	pinFlag := flag.String("pin", "", "six-digit pairing PIN (generated when omitted)")
	udpPort := flag.Int("discovery-port", 8766, "local-network UDP discovery port; set 0 to disable")
	tlsCert := flag.String("tls-cert", "", "TLS certificate PEM file; enables HTTPS/WSS when used with -tls-key")
	tlsKey := flag.String("tls-key", "", "TLS private key PEM file")
	dryRun := flag.Bool("dry-run", false, "log input instead of injecting it (for UI/protocol testing)")
	trustStorePath := flag.String("trust-store", "", "path to trusted-device token database (default: user config directory)")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("TouchBridge %s\n", appVersion)
		return
	}

	if (*tlsCert == "") != (*tlsKey == "") {
		log.Fatal("-tls-cert and -tls-key must be supplied together")
	}
	pin := strings.TrimSpace(*pinFlag)
	if pin != "" && !sixDigits(pin) {
		log.Fatal("-pin must contain exactly six digits")
	}
	port, err := portOf(*listen)
	if err != nil {
		log.Fatalf("invalid TCP listen address %q: %v", *listen, err)
	}

	var cert tls.Certificate
	if *tlsCert != "" {
		cert, err = tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("load TLS certificate/key: %v", err)
		}
	}

	// Reserve TCP before creating the input device or starting UDP discovery.
	// This makes repeated launches harmless and prevents a duplicate Linux
	// process from creating a second virtual input device before it fails.
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			if status, ok := probeExistingInstance(*listen, *tlsCert != ""); ok {
				printExistingInstance(status, *listen, *tlsCert != "")
				return
			}
			log.Fatalf("cannot start: TCP address %s is already in use by another process; inspect it with `ss -ltnp`", *listen)
		}
		log.Fatalf("listen on %s: %v", *listen, err)
	}

	trustStore, err := web.OpenTrustStore(*trustStorePath)
	if err != nil {
		_ = listener.Close()
		log.Fatalf("trusted-device store: %v", err)
	}

	scheme := "http"
	if *tlsCert != "" {
		scheme = "https"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if *udpPort < 0 || *udpPort > 65535 {
		_ = listener.Close()
		log.Fatal("-discovery-port must be between 0 and 65535")
	}
	if *udpPort > 0 {
		if err := discovery.Start(ctx, *udpPort, port, trustStore.ServerID(), scheme); err != nil {
			log.Printf("LAN discovery unavailable: %v (manual IP entry remains possible)", err)
		} else {
			log.Printf("LAN discovery listening on UDP %d", *udpPort)
		}
	}

	var backend input.Backend
	if *dryRun {
		backend = &input.Mock{Logger: log.New(os.Stdout, "", log.LstdFlags)}
	} else {
		backend, err = input.New()
		if err != nil {
			cancel()
			_ = listener.Close()
			log.Fatalf("input backend: %v", err)
		}
	}

	if pin == "" {
		pin, err = web.GeneratePIN()
		if err != nil {
			cancel()
			_ = listener.Close()
			_ = backend.Close()
			log.Fatalf("generate pairing PIN: %v", err)
		}
	}
	pinExpiry := time.Now().Add(web.PairingTTL)
	app := &web.Server{
		PIN: pin, PINExpires: pinExpiry, Version: appVersion, TrustStore: trustStore,
		Input: backend, Logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	srv := &http.Server{
		Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 16,
	}
	if *tlsCert != "" {
		listener = tls.NewListener(listener, &tls.Config{
			Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12,
		})
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(listener) }()
	go reportPINRenewals(ctx, app, log.New(os.Stdout, "", log.LstdFlags))

	// net.Listen succeeded synchronously, so the port is bound before we claim
	// readiness. A Serve error is still monitored and reported below.
	fmt.Println("\nTouchBridge is ready")
	fmt.Printf("  Controller URL : %s://<computer-ip>:%d\n", scheme, port)
	fmt.Printf("  Pairing PIN    : %s  (valid for 10 minutes; renews automatically)\n", pin)
	fmt.Println("  Android app    : Open TouchBridge → Find computers → select this PC")
	fmt.Printf("  Listening      : %s\n", *listen)
	if *dryRun {
		fmt.Println("  Input mode     : DRY RUN (no desktop input will be sent)")
	}
	if *tlsCert == "" {
		fmt.Println("  Security       : HTTP is unencrypted; use only on a trusted private network")
	} else {
		fmt.Println("  Security       : HTTPS/WSS enabled")
	}
	fmt.Println("Press Ctrl+C to stop.")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	select {
	case sig := <-signals:
		log.Printf("received %s; shutting down", sig)
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server stopped unexpectedly: %v", err)
		}
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful HTTP shutdown: %v", err)
		_ = srv.Close()
	}
	if err := app.Shutdown(); err != nil {
		log.Printf("close input backend: %v", err)
	}
}

func sixDigits(value string) bool {
	if len(value) != 6 {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func portOf(addr string) (int, error) {
	_, rawPort, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}
	return port, nil
}

type existingStatus struct {
	Server    string    `json:"server"`
	Version   string    `json:"version"`
	PIN       string    `json:"pin"`
	ExpiresAt time.Time `json:"expires_at"`
}

// probeExistingInstance only queries loopback. The status endpoint refuses
// requests from LAN clients because it returns the active pairing PIN.
func probeExistingInstance(listen string, useTLS bool) (existingStatus, bool) {
	var status existingStatus
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return status, false
	}
	if host == "" || host == "0.0.0.0" || host == "::" || strings.EqualFold(host, "localhost") {
		host = "127.0.0.1"
	} else if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		// A listener bound to one external interface cannot be probed through
		// loopback without changing its configured bind behaviour.
		return status, false
	}
	scheme := "http"
	if useTLS {
		scheme = "https"
	}
	url := scheme + "://" + net.JoinHostPort(host, port) + "/__touchbridge/status"
	transport := &http.Transport{}
	if useTLS {
		// The probe destination is strictly loopback and only verifies the
		// local process identity; it is not used for remote TLS connections.
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12} //nolint:gosec
	}
	client := &http.Client{Timeout: 700 * time.Millisecond, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer transport.CloseIdleConnections()
	resp, err := client.Get(url)
	if err != nil {
		return status, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return status, false
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return status, false
	}
	return status, status.Server == "touchbridge" && sixDigits(status.PIN) && time.Now().Before(status.ExpiresAt)
}

func printExistingInstance(status existingStatus, listen string, useTLS bool) {
	port, _ := portOf(listen)
	scheme := "http"
	if useTLS {
		scheme = "https"
	}
	remaining := time.Until(status.ExpiresAt).Round(time.Second)
	if remaining < 0 {
		remaining = 0
	}
	fmt.Printf("\nTouchBridge is already running (version %s). Reusing the existing instance.\n", status.Version)
	fmt.Printf("  Controller URL : %s://<computer-ip>:%d\n", scheme, port)
	fmt.Printf("  Pairing PIN    : %s  (expires in %s)\n", status.PIN, remaining)
	fmt.Println("  No second server was started; the existing process is still serving connections.")
}

// reportPINRenewals keeps the terminal's pairing code useful after the initial
// ten-minute window. CurrentPairing is mutex-protected, so local duplicate-run
// probes and the renewal timer always observe the same active PIN.
func reportPINRenewals(ctx context.Context, app *web.Server, logger *log.Logger) {
	_, expiry, err := app.CurrentPairing(time.Now())
	if err != nil {
		logger.Printf("pairing PIN renewal unavailable: %v", err)
		return
	}
	for {
		delay := time.Until(expiry)
		if delay < 0 {
			delay = 0
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		pin, nextExpiry, pairingErr := app.CurrentPairing(time.Now())
		if pairingErr != nil {
			logger.Printf("pairing PIN renewal failed: %v", pairingErr)
			expiry = time.Now().Add(30 * time.Second)
			continue
		}
		logger.Printf("Pairing PIN renewed: %s (valid for 10 minutes)", pin)
		expiry = nextExpiry
	}
}
