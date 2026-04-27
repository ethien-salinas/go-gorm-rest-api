package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a private type for context keys in this package to avoid collisions.
type contextKey string

// UserIDKey is the context key under which the authenticated user ID is stored.
const UserIDKey contextKey = "userID"

// Auth returns middleware that validates a Bearer JWT in the Authorization header.
// On success it injects the user ID into the request context under [UserIDKey].
// On failure it responds 401 Unauthorized.
func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeUnauthorized(w)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			}, jwt.WithExpirationRequired())

			if err != nil || !token.Valid {
				writeUnauthorized(w)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeUnauthorized(w)
				return
			}

			userID, ok := claims["sub"]
			if !ok {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user ID injected by [Auth] middleware.
// Returns the ID and true if found; returns 0 and false otherwise.
func UserIDFromContext(ctx context.Context) (uint, bool) {
	v := ctx.Value(UserIDKey)
	if v == nil {
		return 0, false
	}
	// jwt.MapClaims encodes numeric values as float64.
	if f, ok := v.(float64); ok {
		return uint(f), true
	}
	return 0, false
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`)) //nolint:errcheck
}
