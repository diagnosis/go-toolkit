package middleware

import (
	"context"
	"net/http"
)

type UserIDKey struct{}

type AuthFunc = func(r *http.Request) (string, error)

func RequireAuth(authFunc AuthFunc) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := authFunc(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}

			ctx := SetUserID(r.Context(), userID) // Use helper
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth tries to extract user but doesn't fail if missing
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

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey{}).(string)
	return userID, ok
}

// SetUserID sets the user ID in context
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey{}, userID)
}
