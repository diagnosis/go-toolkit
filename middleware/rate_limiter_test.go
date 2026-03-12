package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// handler is a simple 200 OK handler for testing
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func newRequest(ip string) *http.Request {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip + ":1234"
	return req
}

// ── RateLimit middleware ──────────────────────────────────────────────────────

func TestRateLimit_AllowsUnderBurst(t *testing.T) {
	handler := RateLimit(rate.Every(time.Second), 5, time.Minute)(okHandler)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRequest("192.168.1.1"))
		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_RejectsOverBurst(t *testing.T) {
	handler := RateLimit(rate.Every(time.Second), 5, time.Minute)(okHandler)

	// exhaust the burst
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRequest("192.168.1.1"))
	}

	// 6th request should be rejected
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, newRequest("192.168.1.1"))

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	handler := RateLimit(rate.Every(time.Second), 1, time.Minute)(okHandler)

	// exhaust burst of 1
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, newRequest("192.168.1.1"))

	// second request should have Retry-After
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, newRequest("192.168.1.1"))

	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
	if w2.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header to be set")
	}
}

func TestRateLimit_DifferentIPsAreIndependent(t *testing.T) {
	handler := RateLimit(rate.Every(time.Second), 2, time.Minute)(okHandler)

	// exhaust IP A
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newRequest("10.0.0.1"))
	}

	// IP A should be rejected
	wA := httptest.NewRecorder()
	handler.ServeHTTP(wA, newRequest("10.0.0.1"))
	if wA.Code != http.StatusTooManyRequests {
		t.Errorf("IP A: expected 429, got %d", wA.Code)
	}

	// IP B should still be allowed
	wB := httptest.NewRecorder()
	handler.ServeHTTP(wB, newRequest("10.0.0.2"))
	if wB.Code != http.StatusOK {
		t.Errorf("IP B: expected 200, got %d", wB.Code)
	}
}

func TestRateLimit_XForwardedFor(t *testing.T) {
	handler := RateLimit(rate.Every(time.Second), 1, time.Minute)(okHandler)

	// exhaust via X-Forwarded-For IP
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.5")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	// same X-Forwarded-For IP should be rate limited
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.5")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for same forwarded IP, got %d", w2.Code)
	}
}

// ── cleanup ───────────────────────────────────────────────────────────────────

func TestCleanup_RemovesStaleClients(t *testing.T) {
	ttl := 50 * time.Millisecond
	l := newRateLimiter(rate.Every(time.Second), 5, ttl)

	// register a client
	l.getLimiter("192.168.1.100")

	l.mu.Lock()
	count := len(l.clients)
	l.mu.Unlock()

	if count != 1 {
		t.Fatalf("expected 1 client, got %d", count)
	}

	// wait for TTL + one cleanup cycle (ttl/2 tick + ttl age)
	time.Sleep(ttl * 3)

	l.mu.Lock()
	count = len(l.clients)
	l.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 clients after cleanup, got %d", count)
	}
}

func TestCleanup_KeepsActiveClients(t *testing.T) {
	ttl := 100 * time.Millisecond
	l := newRateLimiter(rate.Every(time.Second), 5, ttl)

	// keep refreshing lastSeen before TTL expires
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				l.getLimiter("192.168.1.200")
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	time.Sleep(ttl * 2)
	close(done)

	l.mu.Lock()
	count := len(l.clients)
	l.mu.Unlock()

	if count != 1 {
		t.Errorf("expected active client to survive cleanup, got %d clients", count)
	}
}

// ── realIP ────────────────────────────────────────────────────────────────────

func TestRealIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:5000"

	ip := realIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip)
	}
}

func TestRealIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	req.RemoteAddr = "10.0.0.1:1234"

	ip := realIP(req)
	if ip != "203.0.113.5" {
		t.Errorf("expected 203.0.113.5, got %s", ip)
	}
}

func TestRealIP_InvalidXForwardedFor_FallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "not-an-ip")
	req.RemoteAddr = "192.168.1.1:5000"

	ip := realIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected fallback to 192.168.1.1, got %s", ip)
	}
}
