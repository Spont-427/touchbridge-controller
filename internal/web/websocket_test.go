package web

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpgradeRejectsInvalidWebSocketKey(t *testing.T) {
	r := httptest.NewRequest("GET", "/ws", nil)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Sec-WebSocket-Version", "13")
	r.Header.Set("Sec-WebSocket-Key", base64.StdEncoding.EncodeToString([]byte("too-short")))
	if _, err := upgrade(httptest.NewRecorder(), r); err == nil {
		t.Fatal("invalid WebSocket key was accepted")
	}
}

func TestReadMessageRejectsLargeFrameBeforeAllocatingPayload(t *testing.T) {
	// FIN + text, masked + 16-bit length, then 4097 bytes. The mask/body are
	// intentionally absent: the parser must reject based on the header alone.
	frame := []byte{0x81, 0xfe, 0x10, 0x01}
	c := &wsConn{r: bufio.NewReader(bytes.NewReader(frame))}
	if _, _, err := c.readMessage(); err == nil {
		t.Fatal("oversized frame was accepted")
	}
}

func TestReadMessageRejectsNonCanonicalLength(t *testing.T) {
	// 125 encoded using the 16-bit length marker is invalid WebSocket framing.
	frame := []byte{0x81, 0xfe, 0x00, 0x7d}
	c := &wsConn{r: bufio.NewReader(bytes.NewReader(frame))}
	if _, _, err := c.readMessage(); err == nil {
		t.Fatal("non-canonical frame length was accepted")
	}
}

func FuzzReadMessageNeverPanics(f *testing.F) {
	f.Add([]byte{0x81, 0x80, 1, 2, 3, 4})
	f.Add([]byte{0x81, 0xfe, 0x10, 0x01})
	f.Add([]byte{0x88, 0x80, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 8192 {
			t.Skip()
		}
		c := &wsConn{r: bufio.NewReader(bytes.NewReader(raw))}
		_, _, _ = c.readMessage()
	})
}

func TestUpgradeChecksConnectionHeaderTokensExactly(t *testing.T) {
	newRequest := func(connection string) *http.Request {
		r := httptest.NewRequest("GET", "/ws", nil)
		r.Header.Set("Upgrade", "websocket")
		r.Header.Set("Connection", connection)
		r.Header.Set("Sec-WebSocket-Version", "13")
		r.Header.Set("Sec-WebSocket-Key", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef")))
		return r
	}
	if _, err := upgrade(httptest.NewRecorder(), newRequest("not-upgrade")); err != errWSProtocol {
		t.Fatalf("a substring of the Connection token was accepted: %v", err)
	}
	// A comma-separated Connection header is valid; the recorder lacks
	// Hijacker, so reaching that error proves token parsing succeeded.
	if _, err := upgrade(httptest.NewRecorder(), newRequest("keep-alive, Upgrade")); err == nil || err.Error() != "HTTP hijacking unavailable" {
		t.Fatalf("valid Connection token was not recognized: %v", err)
	}
}
