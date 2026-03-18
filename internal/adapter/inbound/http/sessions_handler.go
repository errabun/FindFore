package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"
)

type loginRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	// Support both "login" (new) and "email" (legacy) fields
	login := strings.TrimSpace(req.Login)
	if login == "" {
		login = strings.TrimSpace(req.Email)
	}

	details, token, err := h.sessions.Login(r.Context(), login, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Invalid email/username or password")
		return
	}

	friends := details.Friends
	if friends == nil {
		friends = []int64{}
	}
	events := details.Events
	if events == nil {
		events = []int64{}
	}

	resp := LoginResponse{
		ID:       details.ID,
		Name:     details.Name,
		Phone:    details.Phone,
		Email:    details.Email,
		Username: details.Username,
		Friends:  friends,
		Events:   events,
		Token:    token,
	}

	respondJSON(w, http.StatusOK, resp)
}
