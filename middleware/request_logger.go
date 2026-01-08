package middleware

import (
	"net/http"
	"time"

	"github.com/diagnosis/go-toolkit/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK // Default to 200
	}
	return rw.ResponseWriter.Write(b)
}

func RequestLogger() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			logger.Info(
				r.Context(),
				"http request started",
				"method", r.Method,
				"path", r.URL.Path,
			)
			next.ServeHTTP(wrapped, r)

			// 3. Calculate duration
			duration := time.Since(start)

			// 4. Log completion
			logger.Info(
				r.Context(),
				"http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", duration.String(),
			)

		})
	}
}
