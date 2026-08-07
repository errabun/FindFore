package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// AccessLog writes one structured slog Info line per request.
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		slog.Info("http_request",
			"request_id", RequestIDFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// Recoverer logs panics with stack + request_id and returns a generic 500 JSON body.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			slog.Error("panic",
				"request_id", RequestIDFromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"panic", rec,
				"stack", string(debug.Stack()),
			)

			w.Header().Set("Content-Type", "application/json")
			if id := RequestIDFromContext(r.Context()); id != "" {
				w.Header().Set(RequestIDHeader, id)
			}
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":[{"code":"internal_error","message":"Internal server error"}]}`))
		}()
		next.ServeHTTP(w, r)
	})
}
