package middleware

import (
	"net/http"

	"github.com/diagnosis/go-toolkit/v3/logger"
	"github.com/google/uuid"
)

// CorrelationID returns middleware that ensures every request has a
// correlation ID: it reuses the incoming X-Correlation-ID header or generates
// a new UUID, echoes it in the response header, and stores it in the request
// context for the logger package to pick up.
func CorrelationID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Correlation-ID")
			if id == "" {
				id = uuid.New().String()
			}

			// ✅ Always set response header (moved outside if)
			w.Header().Set("X-Correlation-ID", id)

			ctx := logger.WithCorrelationID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
