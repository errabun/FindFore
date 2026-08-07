package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const RequestIDHeader = "X-Request-ID"

type requestIDKeyType contextKey

const requestIDKey requestIDKeyType = "request_id"

// RequestIDFromContext returns the request id set by RequestID middleware.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// RequestID ensures every request has an id: honor inbound X-Request-ID, else
// derive from X-Cloud-Trace-Context, else generate. Always echoes X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get(RequestIDHeader))
		if id == "" {
			id = strings.TrimSpace(r.Header.Get("X-Request-Id"))
		}
		if id == "" {
			id = traceIDFromCloudTrace(r.Header.Get("X-Cloud-Trace-Context"))
		}
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func traceIDFromCloudTrace(header string) string {
	// Format: TRACE_ID/SPAN_ID;o=TRACE_TRUE
	if header == "" {
		return ""
	}
	traceID, _, _ := strings.Cut(header, "/")
	return strings.TrimSpace(traceID)
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b[:])
}
