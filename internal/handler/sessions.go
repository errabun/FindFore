package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/model"
	"github.com/ericrabun/findfore-go/internal/store"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	req.Email = strings.ToLower(req.Email)

	player, err := h.queries.GetPlayerByEmail(r.Context(), sql.NullString{String: req.Email, Valid: true})
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Invalid email or password")
		return
	}

	if !auth.CheckPassword(req.Password, player.PasswordDigest.String) {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Invalid email or password")
		return
	}

	token, err := auth.GenerateToken(player.ID, h.jwtSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to generate token")
		return
	}

	details, err := store.GetPlayerWithDetails(r.Context(), h.queries, player.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch player details")
		return
	}

	resp := model.LoginResponse{
		ID:       details.ID,
		Name:     details.Name,
		Phone:    details.Phone,
		Email:    details.Email,
		Username: details.Username,
		Friends:  details.Friends,
		Events:   details.Events,
		Token:    token,
	}

	respondJSON(w, http.StatusOK, resp)
}
