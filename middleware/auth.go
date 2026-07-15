package middleware

import (
	"context"
	"net/http"

	"github.com/diagnosis/go-toolkit/v2/apperr"
	"github.com/diagnosis/go-toolkit/v2/logger"
	"github.com/diagnosis/go-toolkit/v2/responder"
)

// userIDKey is the private context key under which the authenticated user ID
// is stored. Use SetUserID and GetUserID to access it.
type userIDKey struct{}

// AuthFunc extracts and validates the caller's identity from a request,
// returning the authenticated user ID or an error when authentication fails.
type AuthFunc = func(r *http.Request) (string, error)

// RequireAuth returns middleware that authenticates every request with
// authFunc. On success the user ID is stored in the request context
// (retrievable via GetUserID); on failure it responds 401 with the standard
// JSON error envelope and does not call the next handler.
func RequireAuth(authFunc AuthFunc) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := authFunc(r)
			if err != nil {
				correlationID, _ := logger.GetCorrelationID(r.Context())
				responder.Error(w, apperr.Unauthorized("Unauthorized", "authentication failed", err), correlationID)
				return
			}

			ctx := SetUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns middleware that tries to authenticate the request with
// authFunc and stores the user ID in the context when it succeeds, but always
// calls the next handler regardless of the outcome.
func OptionalAuth(authFunc AuthFunc) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := authFunc(r)
			if err == nil && userID != "" {
				ctx := SetUserID(r.Context(), userID)
				r = r.WithContext(ctx)
			}
			// Continue regardless
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID returns the authenticated user ID stored in ctx by SetUserID,
// and whether one was present.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey{}).(string)
	return userID, ok
}

// SetUserID returns a copy of ctx carrying the given user ID.
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}
