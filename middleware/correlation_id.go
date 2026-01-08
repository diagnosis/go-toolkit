package middleware

import (
	"net/http"

	"github.com/diagnosis/go-toolkit/logger"
	"github.com/google/uuid"
)

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
