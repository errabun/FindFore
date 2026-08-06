package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ericrabun/findfore-go/internal/auth"
)

type contextKey string

const PlayerIDKey contextKey = "player_id"

// PlayerIDFromContext returns the authenticated player id when present.
func PlayerIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(PlayerIDKey).(int64)
	if !ok || id <= 0 {
		return 0, false
	}
	return id, true
}

// AuthOptional parses a Bearer JWT when present. Invalid/missing tokens continue
// without a player id (for public routes).
func AuthOptional(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, _ := authenticate(r, jwtSecret)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthRequired requires a valid Bearer JWT. Responds 401 otherwise.
func AuthRequired(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, ok := authenticate(r, jwtSecret)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"errors":[{"code":"unauthorized","message":"Authentication required"}]}`))
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authenticate(r *http.Request, jwtSecret string) (context.Context, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return r.Context(), false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return r.Context(), false
	}

	playerID, err := auth.ValidateToken(parts[1], jwtSecret)
	if err != nil {
		return r.Context(), false
	}

	return context.WithValue(r.Context(), PlayerIDKey, playerID), true
}
