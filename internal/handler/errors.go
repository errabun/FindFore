package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ericrabun/findfore-go/internal/model"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, code string, message string) {
	respondJSON(w, status, model.ErrorResponse{
		Errors: []model.ErrorDetail{
			{Code: code, Message: message},
		},
	})
}
