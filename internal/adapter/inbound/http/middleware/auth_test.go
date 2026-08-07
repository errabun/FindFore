package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/auth"
)

type stubVersions struct {
	version int32
	err     error
}

func (s stubVersions) GetTokenVersion(context.Context, int64) (int32, error) {
	return s.version, s.err
}

func TestAuthRequiredRejectsMissingToken(t *testing.T) {
	h := middleware.AuthRequired("test-secret", stubVersions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthRequiredAcceptsValidToken(t *testing.T) {
	secret := "test-secret"
	token, err := auth.GenerateToken(42, 1, secret)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	var gotID int64
	h := middleware.AuthRequired(secret, stubVersions{version: 1})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.PlayerIDFromContext(r.Context())
		if !ok {
			t.Fatalf("expected player id in context")
		}
		gotID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if gotID != 42 {
		t.Fatalf("expected player 42, got %d", gotID)
	}
}

func TestAuthRequiredRejectsStaleTokenVersion(t *testing.T) {
	secret := "test-secret"
	token, err := auth.GenerateToken(42, 0, secret)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	h := middleware.AuthRequired(secret, stubVersions{version: 1})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for stale token version, got %d", rr.Code)
	}
}
