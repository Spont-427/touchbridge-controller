package discovery

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const Probe = "TOUCHBRIDGE_DISCOVER_V1"

// Start listens for local-network discovery probes and replies only to the
// exact TouchBridge probe. Discovery never authenticates a controller.
func Start(ctx context.Context, udpPort, httpPort int, serverIDs ...string) error {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: udpPort}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("listen for LAN discovery on UDP %d: %w", udpPort, err)
	}
	go func() {
		defer conn.Close()
		host, _ := os.Hostname()
		if host == "" {
			host = "TouchBridge computer"
		}
		serverID := ""
		scheme := "http"
		if len(serverIDs) > 0 {
			serverID = serverIDs[0]
		}
		if len(serverIDs) > 1 && (serverIDs[1] == "http" || serverIDs[1] == "https") {
			scheme = serverIDs[1]
		}
		response := []byte("TOUCHBRIDGE_SERVER_V1\n" + strconv.Itoa(httpPort) + "\n" + strings.ReplaceAll(host, "\n", " ") + "\n" + serverID + "\n" + scheme)
		buf := make([]byte, 128)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(time.Second))
			n, remote, readErr := conn.ReadFromUDP(buf)
			if readErr != nil {
				if ne, ok := readErr.(net.Error); ok && ne.Timeout() {
					select {
					case <-ctx.Done():
						return
					default:
						continue
					}
				}
				return
			}
			if string(buf[:n]) != Probe {
				continue
			}
			_, _ = conn.WriteToUDP(response, remote)
		}
	}()
	return nil
}
