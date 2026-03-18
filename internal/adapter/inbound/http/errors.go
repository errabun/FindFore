package httphandler

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Errors []ErrorDetail `json:"errors"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, code string, message string) {
	respondJSON(w, status, ErrorResponse{
		Errors: []ErrorDetail{
			{Code: code, Message: message},
		},
	})
}
