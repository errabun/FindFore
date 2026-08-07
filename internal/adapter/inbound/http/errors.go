package httphandler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Errors    []ErrorDetail `json:"errors"`
	RequestID string        `json:"request_id,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, code string, message string) {
	respondJSON(w, status, ErrorResponse{
		Errors: []ErrorDetail{
			{Code: code, Message: message},
		},
	})
}

// respondErrorWithRequest includes request_id in the JSON body (useful for 5xx support).
func respondErrorWithRequest(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	respondJSON(w, status, ErrorResponse{
		Errors: []ErrorDetail{
			{Code: code, Message: message},
		},
		RequestID: mw.RequestIDFromContext(r.Context()),
	})
}

func logRequestError(r *http.Request, err error, code string) {
	attrs := []any{
		"request_id", mw.RequestIDFromContext(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"code", code,
	}
	if pid, ok := mw.PlayerIDFromContext(r.Context()); ok {
		attrs = append(attrs, "player_id", pid)
	}
	if err != nil {
		attrs = append(attrs, "err", err.Error())
	}
	slog.Error("handler_error", attrs...)
}

// respondInternalError logs the underlying error and returns a generic 500 to the client.
func respondInternalError(w http.ResponseWriter, r *http.Request, err error, clientMessage string) {
	logRequestError(r, err, "internal_error")
	respondErrorWithRequest(w, r, http.StatusInternalServerError, "internal_error", clientMessage)
}

// respondLoggedError logs when status >= 500, then responds (includes request_id for 5xx).
func respondLoggedError(w http.ResponseWriter, r *http.Request, status int, code, message string, err error) {
	if status >= 500 {
		logRequestError(r, err, code)
		respondErrorWithRequest(w, r, status, code, message)
		return
	}
	respondError(w, status, code, message)
}
