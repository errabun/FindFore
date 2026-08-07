package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoginRateLimiterBlocksAfterFailures(t *testing.T) {
	limiter := NewLoginRateLimiter(3, time.Minute)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))

	body := `{"login":"golfer@example.com","password":"wrong"}`
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))
		req.RemoteAddr = "203.0.113.10:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: got %d, want 401", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", rec.Code)
	}
}

func TestLoginRateLimiterClearsOnSuccess(t *testing.T) {
	limiter := NewLoginRateLimiter(2, time.Minute)
	status := http.StatusUnauthorized
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))

	body := `{"login":"ok@example.com","password":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))
	req.RemoteAddr = "198.51.100.1:99"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}

	status = http.StatusOK
	req = httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))
	req.RemoteAddr = "198.51.100.1:99"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	status = http.StatusUnauthorized
	for i := 0; i < 2; i++ {
		req = httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body))
		req.RemoteAddr = "198.51.100.1:99"
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("post-clear attempt %d: got %d, want 401", i+1, rec.Code)
		}
	}
}
