package web

import (
	"crypto/subtle"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	pinTTL         = 10 * time.Minute
	authWindow     = time.Minute
	maxAttemptsIP  = 5
	maxAttemptsAll = 40
)

type attemptWindow struct {
	start time.Time
	count int
}

type authGuard struct {
	mu     sync.Mutex
	perIP  map[string]attemptWindow
	global attemptWindow
}

func (g *authGuard) allow(ip string, now time.Time) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.perIP == nil {
		g.perIP = make(map[string]attemptWindow)
	}
	if now.Sub(g.global.start) >= authWindow || g.global.start.IsZero() {
		g.global = attemptWindow{start: now}
	}
	if now.Sub(g.global.start) < authWindow && g.global.count >= maxAttemptsAll {
		return false
	}
	w := g.perIP[ip]
	if now.Sub(w.start) >= authWindow || w.start.IsZero() {
		w = attemptWindow{start: now}
	}
	if w.count >= maxAttemptsIP {
		return false
	}
	w.count++
	g.perIP[ip] = w
	g.global.count++
	// Keep the map bounded if many distinct LAN clients probe it.
	if len(g.perIP) > 256 {
		for k, v := range g.perIP {
			if now.Sub(v.start) > 2*authWindow {
				delete(g.perIP, k)
			}
		}
	}
	return true
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return strings.Trim(remoteAddr, "[]")
}

func verifyPIN(candidate, expected string) bool {
	if len(candidate) != 6 || len(expected) != 6 {
		return false
	}
	for _, c := range candidate {
		if c < '0' || c > '9' {
			return false
		}
	}
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(expected)) == 1
}
