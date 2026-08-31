package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/diagnosis/go-toolkit/v3/apperr"
	"github.com/diagnosis/go-toolkit/v3/logger"
	"github.com/diagnosis/go-toolkit/v3/responder"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client

	rate  rate.Limit
	burst int
	ttl   time.Duration
}

func newRateLimiter(requestPerSecond rate.Limit, burst int, ttl time.Duration) *rateLimiter {
	l := &rateLimiter{
		clients: make(map[string]*client),
		rate:    requestPerSecond,
		burst:   burst,
		ttl:     ttl,
	}

	// TODO: cleanup goroutine has no stop mechanism — revisit with graceful
	// shutdown (goroutine lifecycle).
	go l.cleanup()
	return l
}

func (l *rateLimiter) cleanup() {
	ticker := time.NewTicker(l.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		l.mu.Lock()
		for ip, c := range l.clients {
			if now.Sub(c.lastSeen) > l.ttl {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *rateLimiter) getLimiter(ip string) *rate.Limiter {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if c, ok := l.clients[ip]; ok {
		c.lastSeen = now
		return c.limiter
	}

	lim := rate.NewLimiter(l.rate, l.burst)
	l.clients[ip] = &client{
		limiter:  lim,
		lastSeen: now,
	}

	return lim
}

// RateLimit returns middleware that rate-limits requests per client IP using
// a token bucket (requestPerSecond refill rate, burst capacity). Idle client
// entries are evicted after ttl. When the limit is exceeded it responds 429
// with the standard JSON error envelope and a Retry-After header.
func RateLimit(requestPerSecond rate.Limit, burst int, ttl time.Duration) func(handler http.Handler) http.Handler {
	rateLimiter := newRateLimiter(requestPerSecond, burst, ttl)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			limiter := rateLimiter.getLimiter(ip)
			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				correlationID, _ := logger.GetCorrelationID(r.Context())
				responder.Error(w, apperr.TooManyRequests("too many requests", "rate limit exceeded"), correlationID)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func realIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(host) != nil {
		return host
	}

	return r.RemoteAddr
}
