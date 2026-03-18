package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type updatePlayerEventRequest struct {
	PlayerID     int64  `json:"player_id"`
	EventID      int64  `json:"event_id"`
	InviteStatus string `json:"invite_status"`
}

func mapPlayerEventToResponse(pe *entity.PlayerEvent) PlayerEventResponse {
	return PlayerEventResponse{
		ID:           pe.ID,
		PlayerID:     pe.PlayerID,
		EventID:      pe.EventID,
		InviteStatus: pe.InviteStatus.String(),
	}
}

func (h *Handler) UpdatePlayerEvent(w http.ResponseWriter, r *http.Request) {
	var req updatePlayerEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	pe, err := h.playerEvents.UpdateStatus(r.Context(), req.PlayerID, req.EventID, req.InviteStatus)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Player event not found")
		return
	}

	respondJSON(w, http.StatusOK, mapPlayerEventToResponse(pe))
}

type joinEventRequest struct {
	PlayerID int64 `json:"player_id"`
	EventID  int64 `json:"event_id"`
}

func (h *Handler) JoinEvent(w http.ResponseWriter, r *http.Request) {
	var req joinEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	pe, err := h.playerEvents.JoinEvent(r.Context(), req.PlayerID, req.EventID)
	if err != nil {
		// Distinguish between conflict and other errors based on message
		if err.Error() == "player is already part of this event" {
			respondError(w, http.StatusConflict, "conflict", "Player is already part of this event")
			return
		}
		if err.Error() == "event is full" {
			respondError(w, http.StatusConflict, "conflict", "Event is full")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to join event")
		return
	}

	respondJSON(w, http.StatusCreated, mapPlayerEventToResponse(pe))
}
