package web

import (
	"testing"
	"time"
)

func TestVerifyPIN(t *testing.T) {
	for _, tc := range []struct {
		candidate string
		want      bool
	}{{"123456", true}, {"12345", false}, {"1234567", false}, {"abcdef", false}, {"123457", false}} {
		if got := verifyPIN(tc.candidate, "123456"); got != tc.want {
			t.Errorf("verifyPIN(%q)=%v, want %v", tc.candidate, got, tc.want)
		}
	}
}

func TestAuthGuardRateLimits(t *testing.T) {
	var g authGuard
	now := time.Now()
	for i := 0; i < maxAttemptsIP; i++ {
		if !g.allow("192.168.1.4", now) {
			t.Fatalf("attempt %d unexpectedly blocked", i+1)
		}
	}
	if g.allow("192.168.1.4", now) {
		t.Fatal("sixth attempt from same IP should be blocked")
	}
	if !g.allow("192.168.1.5", now) {
		t.Fatal("a different LAN client should still be allowed")
	}
	if !g.allow("192.168.1.4", now.Add(authWindow+time.Second)) {
		t.Fatal("rate limit should reset after the window")
	}
}
