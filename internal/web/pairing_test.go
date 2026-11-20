package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCurrentPairingKeepsPINUntilExpiry(t *testing.T) {
	now := time.Now()
	s := &Server{PIN: "123456", PINExpires: now.Add(time.Minute)}
	firstPIN, firstExpiry, err := s.CurrentPairing(now)
	if err != nil {
		t.Fatal(err)
	}
	secondPIN, secondExpiry, err := s.CurrentPairing(now.Add(30 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if firstPIN != secondPIN || !firstExpiry.Equal(secondExpiry) {
		t.Fatal("pairing changed before its expiry")
	}
}

func TestCurrentPairingRotatesExpiredPIN(t *testing.T) {
	now := time.Now()
	s := &Server{PIN: "123456", PINExpires: now.Add(-time.Second)}
	pin, expiry, err := s.CurrentPairing(now)
	if err != nil {
		t.Fatal(err)
	}
	if !verifyPIN(pin, pin) || !expiry.After(now) {
		t.Fatalf("rotated pairing is invalid: pin=%q expiry=%v", pin, expiry)
	}
	if pin == "123456" {
		t.Fatal("expired PIN was reused")
	}
}

func TestLocalStatusReturnsCurrentPairingOnlyToLoopback(t *testing.T) {
	now := time.Now()
	s := &Server{PIN: "123456", PINExpires: now.Add(-time.Second), Version: "test"}
	request := httptest.NewRequest(http.MethodGet, "/__touchbridge/status", nil)
	request.RemoteAddr = "192.168.1.42:50000"
	blocked := httptest.NewRecorder()
	s.Handler().ServeHTTP(blocked, request)
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("remote status code=%d, want 403", blocked.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/__touchbridge/status", nil)
	request.RemoteAddr = "127.0.0.1:50000"
	allowed := httptest.NewRecorder()
	s.Handler().ServeHTTP(allowed, request)
	if allowed.Code != http.StatusOK {
		t.Fatalf("loopback status code=%d, want 200", allowed.Code)
	}
	var status pairingStatus
	if err := json.Unmarshal(allowed.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Server != "touchbridge" || status.Version != "test" || !verifyPIN(status.PIN, status.PIN) || !status.ExpiresAt.After(now) {
		t.Fatalf("unexpected pairing status: %+v", status)
	}
	if status.PIN == "123456" {
		t.Fatal("status endpoint exposed the expired PIN")
	}
}
