package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Spont-427/touchbridge-controller/internal/web"
)

func TestPortOf(t *testing.T) {
	for _, test := range []struct {
		addr string
		want int
		ok   bool
	}{
		{"0.0.0.0:8765", 8765, true},
		{"[::1]:9000", 9000, true},
		{"localhost:65535", 65535, true},
		{"localhost:http", 0, false},
		{"localhost:8765x", 0, false},
		{"8765", 0, false},
	} {
		got, err := portOf(test.addr)
		if (err == nil) != test.ok || (err == nil && got != test.want) {
			t.Errorf("portOf(%q)=(%d,%v), want (%d, ok=%v)", test.addr, got, err, test.want, test.ok)
		}
	}
}

func TestProbeExistingTouchBridge(t *testing.T) {
	server := httptest.NewServer((&web.Server{
		PIN: "023481", PINExpires: time.Now().Add(time.Minute), Version: "test",
	}).Handler())
	defer server.Close()
	listen := strings.TrimPrefix(server.URL, "http://")
	status, ok := probeExistingInstance(listen, false)
	if !ok {
		t.Fatal("expected local TouchBridge instance to be detected")
	}
	if status.PIN != "023481" || status.Version != "test" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestProbeDoesNotTreatUnrelatedHTTPServerAsTouchBridge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()
	listen := strings.TrimPrefix(server.URL, "http://")
	if _, ok := probeExistingInstance(listen, false); ok {
		t.Fatal("unrelated HTTP server was identified as TouchBridge")
	}
}

func TestExistingStatusRequiresUnexpiredPIN(t *testing.T) {
	server := httptest.NewServer((&web.Server{
		PIN: "123456", PINExpires: time.Now().Add(-time.Second), Version: "test",
	}).Handler())
	defer server.Close()
	listen := strings.TrimPrefix(server.URL, "http://")
	status, ok := probeExistingInstance(listen, false)
	if !ok || !sixDigits(status.PIN) || !status.ExpiresAt.After(time.Now()) {
		t.Fatalf("status did not rotate expired PIN: %+v, ok=%v", status, ok)
	}
	if _, err := strconv.Atoi(status.PIN); err != nil {
		t.Fatalf("PIN is not numeric: %v", err)
	}
}
