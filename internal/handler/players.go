package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/model"
	"github.com/ericrabun/findfore-go/internal/store"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@(?:[-a-z0-9]+\.)+[a-z]{2,}$`)

func (h *Handler) ListPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := h.queries.ListPlayers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch players")
		return
	}

	resp := make([]model.PlayerResponse, len(players))
	for i, p := range players {
		details, err := store.GetPlayerWithDetails(r.Context(), h.queries, p.ID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch player details")
			return
		}
		resp[i] = model.PlayerResponse{
			ID:       details.ID,
			Name:     details.Name,
			Phone:    details.Phone,
			Email:    details.Email,
			Username: details.Username,
			Friends:  details.Friends,
			Events:   details.Events,
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

type createPlayerRequest struct {
	Name                 string `json:"name"`
	Phone                string `json:"phone"`
	Email                string `json:"email"`
	Username             string `json:"username"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (h *Handler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req createPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	req.Email = strings.ToLower(req.Email)

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Name can't be blank")
		return
	}
	if req.Phone == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Phone can't be blank")
		return
	}
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Email can't be blank")
		return
	}
	if !emailRegex.MatchString(req.Email) {
		respondError(w, http.StatusBadRequest, "validation_error", "Email is invalid")
		return
	}
	if req.Username == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Username can't be blank")
		return
	}
	if req.Password == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Password can't be blank")
		return
	}
	if req.Password != req.PasswordConfirmation {
		respondError(w, http.StatusBadRequest, "validation_error", "Password confirmation doesn't match Password")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to hash password")
		return
	}

	player, err := h.queries.CreatePlayer(r.Context(), store.CreatePlayerParams{
		Name:           sql.NullString{String: req.Name, Valid: true},
		Phone:          sql.NullString{String: req.Phone, Valid: true},
		Email:          sql.NullString{String: req.Email, Valid: true},
		Username:       sql.NullString{String: req.Username, Valid: true},
		PasswordDigest: sql.NullString{String: hash, Valid: true},
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			respondError(w, http.StatusBadRequest, "validation_error", "Email or username already taken")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create player")
		return
	}

	resp := model.PlayerResponse{
		ID:       player.ID,
		Name:     player.Name.String,
		Phone:    player.Phone.String,
		Email:    player.Email.String,
		Username: player.Username.String,
		Friends:  []int64{},
		Events:   []int64{},
	}

	respondJSON(w, http.StatusCreated, resp)
}
