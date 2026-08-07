package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDHonorsInboundHeader(t *testing.T) {
	var got string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "client-req-123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got != "client-req-123" {
		t.Fatalf("context id = %q, want client-req-123", got)
	}
	if rec.Header().Get(RequestIDHeader) != "client-req-123" {
		t.Fatalf("response header = %q", rec.Header().Get(RequestIDHeader))
	}
}

func TestRequestIDFromCloudTrace(t *testing.T) {
	var got string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = RequestIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Cloud-Trace-Context", "abcdef0123456789/123;o=1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got != "abcdef0123456789" {
		t.Fatalf("got %q", got)
	}
}

func TestRecovererReturnsJSON500(t *testing.T) {
	// Discard panic logs during test
	slog.SetDefault(slog.New(slog.NewTextHandler(ioDiscard{}, nil)))

	h := RequestID(Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/explode", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	if rec.Header().Get(RequestIDHeader) == "" {
		t.Fatal("missing request id header")
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errors, ok := body["errors"].([]any)
	if !ok || len(errors) == 0 {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func TestAccessLogDoesNotBreakHandler(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(ioDiscard{}, nil)))

	h := RequestID(AccessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r.Context()) == "" {
			t.Error("missing request id in handler")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	})))

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("body %q", rec.Body.String())
	}
}
