package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h *Handler) ListPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := h.players.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch players")
		return
	}

	resp := make([]PlayerResponse, len(players))
	for i, p := range players {
		resp[i] = mapPlayerToResponse(p)
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
