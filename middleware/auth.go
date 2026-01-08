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

			ctx := context.WithValue(r.Context(), UserIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey{}).(string)
	return userID, ok
}
