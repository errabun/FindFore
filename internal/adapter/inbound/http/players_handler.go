package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/application/service"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func mapPlayerToResponse(p entity.PlayerWithDetails) PlayerResponse {
	friends := p.Friends
	if friends == nil {
		friends = []int64{}
	}
	events := p.Events
	if events == nil {
		events = []int64{}
	}
	return PlayerResponse{
		ID:       p.ID,
		Name:     p.Name,
		Phone:    p.Phone,
		Email:    p.Email,
		Username: p.Username,
		Friends:  friends,
		Events:   events,
	}
}

// mapPlayerToPublicResponse omits email/phone for community directory listings.
func mapPlayerToPublicResponse(p entity.PlayerWithDetails) PlayerResponse {
	friends := p.Friends
	if friends == nil {
		friends = []int64{}
	}
	events := p.Events
	if events == nil {
		events = []int64{}
	}
	return PlayerResponse{
		ID:       p.ID,
		Name:     p.Name,
		Phone:    "",
		Email:    "",
		Username: p.Username,
		Friends:  friends,
		Events:   events,
	}
}

func (h *Handler) ListPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := h.players.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch players")
		return
	}

	resp := make([]PlayerResponse, len(players))
	for i, p := range players {
		resp[i] = mapPlayerToPublicResponse(p)
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

	player, err := h.players.Create(r.Context(), req.Name, req.Phone, req.Email, req.Username, req.Password, req.PasswordConfirmation)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, "validation_error", ve.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create player")
		return
	}

	resp := PlayerResponse{
		ID:       player.ID,
		Name:     player.Name,
		Phone:    player.Phone,
		Email:    player.Email,
		Username: player.Username,
		Friends:  []int64{},
		Events:   []int64{},
	}

	respondJSON(w, http.StatusCreated, resp)
}

type updatePlayerRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (h *Handler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	callerID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	playerID, err := strconv.ParseInt(chi.URLParam(r, "player_id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid player ID")
		return
	}

	if callerID != playerID {
		respondError(w, http.StatusForbidden, "forbidden", "You can only update your own profile")
		return
	}

	var req updatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	player, err := h.players.Update(r.Context(), callerID, req.Name, req.Phone, req.Email, req.Username)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, "validation_error", ve.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to update player")
		return
	}

	respondJSON(w, http.StatusOK, mapPlayerToResponse(*player))
}

type changePasswordRequest struct {
	CurrentPassword      string `json:"current_password"`
	NewPassword          string `json:"new_password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	callerID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	playerID, err := strconv.ParseInt(chi.URLParam(r, "player_id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid player ID")
		return
	}

	if callerID != playerID {
		respondError(w, http.StatusForbidden, "forbidden", "You can only change your own password")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	if err := h.players.ChangePassword(r.Context(), callerID, req.CurrentPassword, req.NewPassword, req.PasswordConfirmation); err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, "validation_error", ve.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to change password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}
