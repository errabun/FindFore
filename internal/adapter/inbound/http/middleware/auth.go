package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ericrabun/findfore-go/internal/auth"
)

type contextKey string

const PlayerIDKey contextKey = "player_id"

// TokenVersionLookup loads the current token_version for a player (password-change invalidation).
type TokenVersionLookup interface {
	GetTokenVersion(ctx context.Context, playerID int64) (int32, error)
}

// PlayerIDFromContext returns the authenticated player id when present.
func PlayerIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(PlayerIDKey).(int64)
	if !ok || id <= 0 {
		return 0, false
	}
	return id, true
}

// AuthOptional parses a Bearer JWT when present. Invalid/missing tokens continue
// without a player id (for public routes). Does not check token_version.
func AuthOptional(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, ok := parseBearer(r, jwtSecret)
			if ok {
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthRequired requires a valid Bearer JWT whose token_version matches the DB.
func AuthRequired(jwtSecret string, versions TokenVersionLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, ok := parseBearer(r, jwtSecret)
			if !ok {
				writeUnauthorized(w)
				return
			}

			playerID, _ := PlayerIDFromContext(ctx)
			claimsVersion := tokenVersionFromContext(ctx)
			current, err := versions.GetTokenVersion(ctx, playerID)
			if err != nil || current != claimsVersion {
				writeUnauthorized(w)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type tokenVersionKeyType contextKey

const tokenVersionKey tokenVersionKeyType = "token_version"

func tokenVersionFromContext(ctx context.Context) int32 {
	v, _ := ctx.Value(tokenVersionKey).(int32)
	return v
}

func parseBearer(r *http.Request, jwtSecret string) (context.Context, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return r.Context(), false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return r.Context(), false
	}

	claims, err := auth.ValidateToken(parts[1], jwtSecret)
	if err != nil {
		return r.Context(), false
	}

	ctx := context.WithValue(r.Context(), PlayerIDKey, claims.PlayerID)
	ctx = context.WithValue(ctx, tokenVersionKey, claims.TokenVersion)
	return ctx, true
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"code":"unauthorized","message":"Authentication required"}]}`))
}
