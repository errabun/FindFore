package middleware

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LoginRateLimiter limits failed POST /sessions attempts per client IP (+ login when present).
// In-memory only — suitable for a single Cloud Run instance early on.
type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewLoginRateLimiter(limit int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (l *LoginRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)

		body, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			var payload struct {
				Login string `json:"login"`
				Email string `json:"email"`
			}
			if json.Unmarshal(body, &payload) == nil {
				ident := strings.ToLower(strings.TrimSpace(payload.Login))
				if ident == "" {
					ident = strings.ToLower(strings.TrimSpace(payload.Email))
				}
				if ident != "" {
					key = key + "|" + ident
				}
			}
		}

		if l.tooMany(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"errors":[{"code":"rate_limited","message":"Too many login attempts. Please try again later."}]}`))
			return
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		if rec.status == http.StatusUnauthorized {
			l.record(key)
		} else if rec.status >= 200 && rec.status < 300 {
			l.clear(key)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (l *LoginRateLimiter) tooMany(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(key)
	return len(l.attempts[key]) >= l.limit
}

func (l *LoginRateLimiter) record(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(key)
	l.attempts[key] = append(l.attempts[key], time.Now())
}

func (l *LoginRateLimiter) clear(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

func (l *LoginRateLimiter) pruneLocked(key string) {
	cutoff := time.Now().Add(-l.window)
	kept := l.attempts[key][:0]
	for _, t := range l.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.attempts, key)
	} else {
		l.attempts[key] = kept
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
