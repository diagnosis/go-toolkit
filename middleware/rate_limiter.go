package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
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
func RateLimit(requestPerSecond rate.Limit, burst int, ttl time.Duration) func(handler http.Handler) http.Handler {
	rateLimiter := newRateLimiter(requestPerSecond, burst, ttl)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			limiter := rateLimiter.getLimiter(ip)
			if !limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"too many requests"}`))
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
