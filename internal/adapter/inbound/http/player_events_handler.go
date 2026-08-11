package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type updatePlayerEventRequest struct {
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
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req updatePlayerEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	pe, err := h.playerEvents.UpdateStatus(r.Context(), actorID, req.EventID, req.InviteStatus)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Player event not found")
		return
	}

	respondJSON(w, http.StatusOK, mapPlayerEventToResponse(pe))
}

type joinEventRequest struct {
	EventID int64 `json:"event_id"`
}

func (h *Handler) JoinEvent(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req joinEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	pe, err := h.playerEvents.JoinEvent(r.Context(), actorID, req.EventID)
	if err != nil {
		if errors.Is(err, entity.ErrAlreadyOnEvent) {
			respondError(w, http.StatusConflict, "conflict", "Player is already part of this event")
			return
		}
		if errors.Is(err, entity.ErrEventFull) {
			respondError(w, http.StatusConflict, "conflict", "Event is full")
			return
		}
		if errors.Is(err, entity.ErrEventMissing) {
			respondError(w, http.StatusNotFound, "not_found", "Event not found")
			return
		}
		respondInternalError(w, r, err, "Failed to join event")
		return
	}

	respondJSON(w, http.StatusCreated, mapPlayerEventToResponse(pe))
}
