package web

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type testBackend struct {
	moves    atomic.Int32
	releases atomic.Int32
	closed   atomic.Bool
}

func (b *testBackend) MoveMouse(dx, dy int) error   { b.moves.Add(1); return nil }
func (*testBackend) MouseButton(string, bool) error { return nil }
func (*testBackend) Scroll(int) error               { return nil }
func (*testBackend) Key(string, bool) error         { return nil }
func (b *testBackend) ReleaseAll() error            { b.releases.Add(1); return nil }
func (b *testBackend) Close() error                 { b.closed.Store(true); return nil }

type testWS struct {
	conn   net.Conn
	reader *bufio.Reader
}

func dialTestWS(t *testing.T, serverURL, origin string) *testWS {
	t.Helper()
	addr := strings.TrimPrefix(strings.TrimPrefix(serverURL, "http://"), "https://")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))
	_, err = fmt.Fprintf(conn, "GET /ws HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\nOrigin: %s\r\n\r\n", addr, key, origin)
	if err != nil {
		conn.Close()
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, &http.Request{Method: "GET"})
	if err != nil {
		conn.Close()
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		t.Fatalf("upgrade status=%d", resp.StatusCode)
	}
	return &testWS{conn: conn, reader: reader}
}
func (c *testWS) sendJSON(t *testing.T, v any) {
	t.Helper()
	payload, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	mask := []byte{0x11, 0x22, 0x33, 0x44}
	header := []byte{0x81}
	switch n := len(payload); {
	case n < 126:
		header = append(header, 0x80|byte(n))
	case n <= 65535:
		header = append(header, 0x80|126, byte(n>>8), byte(n))
	default:
		t.Fatal("test frame too long")
	}
	header = append(header, mask...)
	for i := range payload {
		payload[i] ^= mask[i&3]
	}
	if _, err = c.conn.Write(append(header, payload...)); err != nil {
		t.Fatal(err)
	}
}
func (c *testWS) readJSON(t *testing.T) map[string]any {
	t.Helper()
	var h [2]byte
	if _, err := io.ReadFull(c.reader, h[:]); err != nil {
		t.Fatal(err)
	}
	if h[0]&0x0f != 1 {
		t.Fatalf("expected text frame, opcode=%d", h[0]&0x0f)
	}
	n := uint64(h[1] & 0x7f)
	if n == 126 {
		var ext [2]byte
		if _, err := io.ReadFull(c.reader, ext[:]); err != nil {
			t.Fatal(err)
		}
		n = uint64(binary.BigEndian.Uint16(ext[:]))
	}
	payload := make([]byte, n)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(payload, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func (c *testWS) close() { _ = c.conn.Close() }

func TestHandlerServesClient(t *testing.T) {
	backend := &testBackend{}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()
	for path, want := range map[string]string{"/healthz": "ok", "/": "GAME CONTROLLER", "/app.js": "WebSocket"} {
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != 200 || !strings.Contains(string(body), want) {
			t.Fatalf("GET %s: status=%d body missing %q", path, resp.StatusCode, want)
		}
	}
}

func TestWebSocketAuthenticationAndInput(t *testing.T) {
	backend := &testBackend{}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Version: "0.6.0", Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()
	c := dialTestWS(t, server.URL, server.URL)
	c.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	reply := c.readJSON(t)
	if reply["type"] != "auth_ok" {
		t.Fatalf("auth response=%v", reply)
	}
	if reply["server_version"] != "0.6.0" {
		t.Fatalf("server_version=%v, want 0.6.0", reply["server_version"])
	}
	c.sendJSON(t, map[string]any{"type": "mouse_move", "dx": 14, "dy": -3})
	deadline := time.Now().Add(time.Second)
	for backend.moves.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if backend.moves.Load() != 1 {
		t.Fatalf("mouse event count=%d", backend.moves.Load())
	}
	c.close()
	deadline = time.Now().Add(time.Second)
	for backend.releases.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if backend.releases.Load() == 0 {
		t.Fatal("disconnect did not release held inputs")
	}
}

func TestWebSocketRejectsWrongPINAndReleasesSlot(t *testing.T) {
	backend := &testBackend{}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()

	bad := dialTestWS(t, server.URL, server.URL)
	bad.sendJSON(t, map[string]any{"type": "auth", "pin": "999999"})
	if got := bad.readJSON(t)["type"]; got != "auth_error" {
		t.Fatalf("wrong PIN response=%v", got)
	}
	bad.close()

	good := dialTestWS(t, server.URL, server.URL)
	defer good.close()
	good.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := good.readJSON(t)["type"]; got != "auth_ok" {
		t.Fatalf("valid PIN after a failed attempt response=%v", got)
	}
}

func TestUnauthenticatedSocketDoesNotBlockValidController(t *testing.T) {
	backend := &testBackend{}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()

	// Keep this first WebSocket open but do not send an auth frame.
	unauthenticated := dialTestWS(t, server.URL, server.URL)
	defer unauthenticated.close()

	valid := dialTestWS(t, server.URL, server.URL)
	defer valid.close()
	valid.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := valid.readJSON(t)["type"]; got != "auth_ok" {
		t.Fatalf("valid controller was blocked by unauthenticated socket: %v", got)
	}
}

func TestOnlyOneAuthenticatedControllerCanConnect(t *testing.T) {
	backend := &testBackend{}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()

	first := dialTestWS(t, server.URL, server.URL)
	defer first.close()
	first.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := first.readJSON(t)["type"]; got != "auth_ok" {
		t.Fatalf("first auth response=%v", got)
	}

	second := dialTestWS(t, server.URL, server.URL)
	defer second.close()
	second.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := second.readJSON(t)["type"]; got != "controller_busy" {
		t.Fatalf("second auth response=%v, want controller_busy", got)
	}
}

func TestWebSocketRejectsWrongOrigin(t *testing.T) {
	backend := &testBackend{}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()
	addr := strings.TrimPrefix(server.URL, "http://")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))
	_, err = fmt.Fprintf(conn, "GET /ws HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\nOrigin: http://attacker.invalid\r\n\r\n", addr, key)
	if err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(line, "403") {
		t.Fatalf("expected 403, got %q", line)
	}
}

func TestWebSocketEnforcesPINRateLimit(t *testing.T) {
	backend := &testBackend{}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()

	for attempt := 1; attempt <= maxAttemptsIP; attempt++ {
		c := dialTestWS(t, server.URL, server.URL)
		c.sendJSON(t, map[string]any{"type": "auth", "pin": "999999"})
		if got := c.readJSON(t)["type"]; got != "auth_error" {
			t.Fatalf("attempt %d response=%v, want auth_error", attempt, got)
		}
		c.close()
	}

	limited := dialTestWS(t, server.URL, server.URL)
	limited.sendJSON(t, map[string]any{"type": "auth", "pin": "999999"})
	if got := limited.readJSON(t)["type"]; got != "rate_limited" {
		t.Fatalf("sixth attempt response=%v, want rate_limited", got)
	}
	limited.close()

	// A correct PIN is not blocked by the failed-attempt limiter.
	good := dialTestWS(t, server.URL, server.URL)
	defer good.close()
	good.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := good.readJSON(t)["type"]; got != "auth_ok" {
		t.Fatalf("valid PIN after rate limit response=%v, want auth_ok", got)
	}
}

func TestWebSocketKeepaliveAndIdleRelease(t *testing.T) {
	backend := &testBackend{}
	app := &Server{
		PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend,
		Logger: log.New(io.Discard, "", 0), IdleTimeout: 250 * time.Millisecond,
	}
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	controller := dialTestWS(t, server.URL, server.URL)
	controller.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := controller.readJSON(t)["type"]; got != "auth_ok" {
		t.Fatalf("auth response=%v", got)
	}
	if got := controller.readJSON(t)["type"]; got != "capabilities" {
		t.Fatalf("capabilities response=%v", got)
	}
	controller.sendJSON(t, map[string]any{"type": "keepalive"})
	if got := controller.readJSON(t)["type"]; got != "keepalive_ack" {
		t.Fatalf("keepalive response=%v", got)
	}

	deadline := time.Now().Add(time.Second)
	for backend.releases.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if backend.releases.Load() == 0 {
		t.Fatal("idle controller connection did not time out and release held input")
	}

	// The connection timeout must release the single-controller slot.
	next := dialTestWS(t, server.URL, server.URL)
	defer next.close()
	next.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := next.readJSON(t)["type"]; got != "auth_ok" {
		t.Fatalf("new controller after idle timeout got %v", got)
	}
}

func TestShutdownClosesHijackedWebSocketBeforeBackend(t *testing.T) {
	backend := &testBackend{}
	app := &Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), Input: backend, Logger: log.New(io.Discard, "", 0)}
	server := httptest.NewServer(app.Handler())
	controller := dialTestWS(t, server.URL, server.URL)
	controller.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if got := controller.readJSON(t)["type"]; got != "auth_ok" {
		t.Fatalf("auth response=%v", got)
	}
	if got := controller.readJSON(t)["type"]; got != "capabilities" {
		t.Fatalf("capabilities response=%v", got)
	}

	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- app.Shutdown() }()

	_ = controller.conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := controller.reader.ReadByte(); err == nil {
		t.Fatal("active WebSocket remained open after application shutdown")
	}
	select {
	case err := <-shutdownDone:
		if err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not wait for WebSocket handler to exit")
	}
	if !backend.closed.Load() {
		t.Fatal("input backend was not closed after WebSocket handler exited")
	}
	if backend.releases.Load() == 0 {
		t.Fatal("active input was not released before backend close")
	}
	server.Close()
}

func TestWebSocketConnectionLimit(t *testing.T) {
	app := &Server{}
	connections := make([]*wsConn, 0, maxWebSocketConnections)
	for i := 0; i < maxWebSocketConnections; i++ {
		c := &wsConn{}
		if !app.registerWebSocket(c) {
			t.Fatalf("connection %d unexpectedly rejected", i+1)
		}
		connections = append(connections, c)
	}
	if app.registerWebSocket(&wsConn{}) {
		t.Fatal("connection beyond the configured limit was accepted")
	}
	for _, c := range connections {
		app.unregisterWebSocket(c)
	}
}

func TestTrustedDeviceCanReconnectAfterPINChanges(t *testing.T) {
	backend := &testBackend{}
	store, err := OpenTrustStore(t.TempDir() + "/trusted.json")
	if err != nil {
		t.Fatal(err)
	}
	app := &Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), TrustStore: store, Input: backend, Logger: log.New(io.Discard, "", 0)}
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	paired := dialTestWS(t, server.URL, server.URL)
	paired.sendJSON(t, map[string]any{"type": "auth", "pin": "123456", "remember_device": true, "device_name": "Test phone"})
	pairReply := paired.readJSON(t)
	paired.close()
	if pairReply["type"] != "auth_ok" {
		t.Fatalf("pair response=%v", pairReply)
	}
	token, ok := pairReply["device_token"].(string)
	if !ok || len(token) != 64 {
		t.Fatalf("expected 64-character device token, got %v", pairReply["device_token"])
	}
	if pairReply["server_id"] != store.ServerID() {
		t.Fatalf("server ID mismatch: %v", pairReply["server_id"])
	}

	app.pinMu.Lock()
	app.PIN = "654321"
	app.PINExpires = time.Now().Add(time.Hour)
	app.pinMu.Unlock()

	reconnected := dialTestWS(t, server.URL, server.URL)
	defer reconnected.close()
	reconnected.sendJSON(t, map[string]any{"type": "auth_device", "device_token": token})
	if reply := reconnected.readJSON(t); reply["type"] != "auth_ok" {
		t.Fatalf("trusted reconnect response=%v", reply)
	}
}

func TestNativeAndroidAssetOriginIsAllowed(t *testing.T) {
	backend := &testBackend{}
	store, err := OpenTrustStore(t.TempDir() + "/trusted.json")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer((&Server{PIN: "123456", PINExpires: time.Now().Add(time.Hour), TrustStore: store, Input: backend, Logger: log.New(io.Discard, "", 0)}).Handler())
	defer server.Close()
	c := dialTestWS(t, server.URL, "https://touchbridge.appassets.local")
	defer c.close()
	c.sendJSON(t, map[string]any{"type": "auth", "pin": "123456"})
	if reply := c.readJSON(t); reply["type"] != "auth_ok" {
		t.Fatalf("native Android origin was rejected: %v", reply)
	}
}
