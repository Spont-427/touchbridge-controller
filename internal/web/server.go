package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Spont-427/touchbridge-controller/internal/input"
	"github.com/Spont-427/touchbridge-controller/internal/protocol"
)

//go:embed all:web
var assets embed.FS

type Server struct {
	PIN          string
	PINExpires   time.Time
	TrustStore   *TrustStore
	Version      string
	IdleTimeout  time.Duration
	Input        input.Backend
	Logger       *log.Logger
	closing      atomic.Bool
	guard        authGuard
	pinMu        sync.Mutex
	inputMu      sync.Mutex
	inputOwners  *inputOwnership
	slotsMu      sync.Mutex
	slots        [input.MaxGamepadSlots]bool
	wsMu         sync.Mutex
	wsConns      map[*wsConn]struct{}
	wsWG         sync.WaitGroup
	shutdownOnce sync.Once
	shutdownErr  error
}

const maxWebSocketConnections = 8

type pairingStatus struct {
	Server    string    `json:"server"`
	Version   string    `json:"version"`
	PIN       string    `json:"pin"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CurrentPairing returns the active pairing PIN and rotates it when its
// lifetime has elapsed. Callers receive a consistent PIN/expiry pair.
func (s *Server) CurrentPairing(now time.Time) (string, time.Time, error) {
	s.pinMu.Lock()
	defer s.pinMu.Unlock()
	if s.PIN == "" || !now.Before(s.PINExpires) {
		pin, err := GeneratePIN()
		if err != nil {
			return "", time.Time{}, err
		}
		s.PIN = pin
		s.PINExpires = now.Add(PairingTTL)
	}
	return s.PIN, s.PINExpires, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	webFS, _ := fs.Sub(assets, "web")
	mux.Handle("/", http.FileServer(http.FS(webFS)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/__touchbridge/status", s.handleLocalStatus)
	mux.HandleFunc("/ws", s.handleWS)
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// handleLocalStatus is intentionally loopback-only. The launcher uses it to
// distinguish an existing TouchBridge instance from an unrelated process
// occupying the port and to display that instance's current PIN.
func (s *Server) handleLocalStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ip := net.ParseIP(remoteIP(r.RemoteAddr))
	if ip == nil || !ip.IsLoopback() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	pin, expiry, err := s.CurrentPairing(time.Now())
	if err != nil {
		http.Error(w, "pairing status unavailable", http.StatusInternalServerError)
		return
	}
	version := s.Version
	if version == "" {
		version = "unknown"
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_ = json.NewEncoder(w).Encode(pairingStatus{Server: "touchbridge", Version: version, PIN: pin, ExpiresAt: expiry})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		expectedScheme := "http"
		if r.TLS != nil {
			expectedScheme = "https"
		}
		nativeAndroidOrigin := strings.EqualFold(origin, "https://touchbridge.appassets.local")
		if err != nil || (!nativeAndroidOrigin && (!strings.EqualFold(u.Host, r.Host) || !strings.EqualFold(u.Scheme, expectedScheme))) {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
	}
	c, err := upgrade(w, r)
	if err != nil {
		s.Logger.Printf("websocket upgrade: %v", err)
		return
	}
	if !s.registerWebSocket(c) {
		_ = c.conn.Close()
		return
	}
	defer func() {
		s.unregisterWebSocket(c)
		_ = c.close()
	}()
	_ = c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	op, payload, err := c.readMessage()
	if err != nil || op != 1 {
		s.Logger.Printf("client disconnected before authentication")
		return
	}
	m, err := protocol.Decode(payload)
	validMessage := err == nil && (m.Type == "auth" || m.Type == "auth_device")
	validAuth := false
	if validMessage && m.Type == "auth" {
		currentPIN, expiry, pairingErr := s.CurrentPairing(time.Now())
		validAuth = pairingErr == nil && time.Now().Before(expiry) && verifyPIN(m.PIN, currentPIN)
	} else if validMessage && m.Type == "auth_device" {
		validAuth = s.TrustStore != nil && s.TrustStore.Verify(m.DeviceToken)
	}
	if !validAuth {
		if validMessage && !s.guard.allow(remoteIP(r.RemoteAddr), time.Now()) {
			_ = c.writeFrame(1, []byte(`{"type":"rate_limited"}`))
			s.Logger.Printf("rate-limited authentication attempts from %s", remoteIP(r.RemoteAddr))
			return
		}
		_ = c.writeFrame(1, []byte(`{"type":"auth_error"}`))
		s.Logger.Printf("rejected unauthenticated controller")
		return
	}
	// Reserve a controller slot only after authentication. Otherwise an
	// unauthenticated client could occupy a player slot during the 10-second
	// auth timeout and block a legitimate controller on the LAN.
	limit := s.controllerLimit()
	slot, ok := s.acquireControllerSlot(limit)
	if !ok {
		busy, _ := json.Marshal(map[string]any{"type": "controller_busy", "max_controllers": limit})
		_ = c.writeFrame(1, busy)
		s.Logger.Printf("rejected controller: all %d player slots are in use", limit)
		return
	}
	s.inputMu.Lock()
	if s.inputOwners == nil {
		s.inputOwners = newInputOwnership(s.Input)
	}
	session := newControllerSession(s.Input, slot, s.inputOwners)
	s.inputMu.Unlock()
	defer func() {
		s.inputMu.Lock()
		if err := session.ReleaseAll(); err != nil {
			s.Logger.Printf("could not release all input from player %d: %v", slot+1, err)
		}
		s.inputMu.Unlock()
		s.releaseControllerSlot(slot)
	}()
	_ = c.conn.SetReadDeadline(time.Time{})
	version := s.Version
	if version == "" {
		version = "unknown"
	}
	authReply := map[string]any{
		"type": "auth_ok", "server_version": version,
		"player": slot + 1, "player_index": slot, "max_controllers": limit,
	}
	if s.TrustStore != nil {
		authReply["server_id"] = s.TrustStore.ServerID()
	}
	// Only PIN pairing can mint a new persistent device credential. Trusted
	// reconnects use the existing bearer token and do not create new records.
	if m.Type == "auth" && m.Remember && s.TrustStore != nil {
		token, serverID, trustErr := s.TrustStore.Issue(m.DeviceName, time.Now())
		if trustErr != nil {
			s.Logger.Printf("could not persist trusted device: %v", trustErr)
		} else {
			authReply["device_token"] = token
			authReply["server_id"] = serverID
		}
	}
	encodedAuthReply, _ := json.Marshal(authReply)
	if err := c.writeFrame(1, encodedAuthReply); err != nil {
		return
	}
	gamepadAvailable := false
	gamepadStatus := "Virtual gamepad is not available on this platform. Mouse and keyboard remain available."
	gamepadModel := "standard-gamepad"
	// Lazy per-player gamepad creation mutates backend state, so capability
	// probes must share the same lock as input handling.
	s.inputMu.Lock()
	if gamepad, ok := s.Input.(input.GamepadBackend); ok {
		gamepadAvailable = gamepad.GamepadAvailable()
	}
	if provider, ok := s.Input.(input.GamepadStatusProvider); ok {
		if status := strings.TrimSpace(provider.GamepadStatus()); status != "" {
			gamepadStatus = status
		}
	}
	s.inputMu.Unlock()
	capabilityMessage, _ := json.Marshal(map[string]any{
		"type": "capabilities", "gamepad": gamepadAvailable,
		"gamepad_model": gamepadModel, "gamepad_status": gamepadStatus,
		"latency_probe": true, "player": slot + 1,
		"player_index": slot, "max_controllers": limit,
	})
	if err := c.writeFrame(1, capabilityMessage); err != nil {
		return
	}
	remote := r.RemoteAddr
	s.Logger.Printf("controller connected: %s (player %d/%d)", remote, slot+1, limit)
	defer s.Logger.Printf("controller disconnected: %s (player %d/%d)", remote, slot+1, limit)
	idleTimeout := s.IdleTimeout
	if idleTimeout <= 0 {
		idleTimeout = 45 * time.Second
	}
	for {
		_ = c.conn.SetReadDeadline(time.Now().Add(idleTimeout))
		op, payload, err = c.readMessage()
		if err != nil {
			return
		}
		switch op {
		case 8:
			return
		case 9:
			if c.writeFrame(10, payload) != nil {
				return
			}
			continue
		case 10:
			continue
		case 1:
			m, err = protocol.Decode(payload)
			if err != nil {
				continue
			}
			if m.Type == "keepalive" {
				if c.writeFrame(1, []byte(`{"type":"keepalive_ack"}`)) != nil {
					return
				}
				continue
			}
			if m.Type == "latency_ping" {
				pong, _ := json.Marshal(map[string]any{"type": "latency_pong", "timestamp": m.Timestamp})
				if c.writeFrame(1, pong) != nil {
					return
				}
				continue
			}
			s.inputMu.Lock()
			err = protocol.Apply(m, session)
			s.inputMu.Unlock()
			if err != nil {
				s.Logger.Printf("ignored input from player %d: %v", slot+1, err)
			}
		default:
			return
		}
	}
}

func (s *Server) controllerLimit() int {
	limit := 1
	if multi, ok := s.Input.(input.GamepadSlotBackend); ok {
		if count := multi.GamepadSlotCount(); count > 0 {
			limit = count
		}
	}
	if limit > input.MaxGamepadSlots {
		limit = input.MaxGamepadSlots
	}
	return limit
}

func (s *Server) acquireControllerSlot(limit int) (int, bool) {
	s.slotsMu.Lock()
	defer s.slotsMu.Unlock()
	for slot := 0; slot < limit && slot < len(s.slots); slot++ {
		if !s.slots[slot] {
			s.slots[slot] = true
			return slot, true
		}
	}
	return 0, false
}

func (s *Server) releaseControllerSlot(slot int) {
	s.slotsMu.Lock()
	defer s.slotsMu.Unlock()
	if slot >= 0 && slot < len(s.slots) {
		s.slots[slot] = false
	}
}

func (s *Server) registerWebSocket(c *wsConn) bool {
	s.wsMu.Lock()
	defer s.wsMu.Unlock()
	if s.closing.Load() {
		return false
	}
	if s.wsConns == nil {
		s.wsConns = make(map[*wsConn]struct{})
	}
	// Bound unauthenticated and active sockets so a LAN peer cannot exhaust
	// file descriptors/memory by opening an arbitrary number of idle upgrades.
	if len(s.wsConns) >= maxWebSocketConnections {
		return false
	}
	s.wsConns[c] = struct{}{}
	s.wsWG.Add(1)
	return true
}

func (s *Server) unregisterWebSocket(c *wsConn) {
	s.wsMu.Lock()
	delete(s.wsConns, c)
	s.wsMu.Unlock()
	s.wsWG.Done()
}

// Shutdown closes hijacked WebSocket connections as well as the input backend.
// net/http.Server.Shutdown does not manage connections after WebSocket upgrade;
// waiting here prevents an input handler racing with backend/device teardown.
func (s *Server) Shutdown() error {
	s.shutdownOnce.Do(func() {
		s.wsMu.Lock()
		s.closing.Store(true)
		connections := make([]*wsConn, 0, len(s.wsConns))
		for c := range s.wsConns {
			connections = append(connections, c)
		}
		s.wsMu.Unlock()

		for _, c := range connections {
			_ = c.conn.Close()
		}
		s.wsWG.Wait()
		if s.Input != nil {
			s.shutdownErr = s.Input.Close()
		}
	})
	return s.shutdownErr
}
