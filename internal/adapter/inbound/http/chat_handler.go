package httphandler

import (
	"errors"
	"net/http"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/application/chat"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type groupChatResponse struct {
	APIKey      string `json:"api_key"`
	Token       string `json:"token"`
	ChannelType string `json:"channel_type"`
	ChannelID   string `json:"channel_id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
}

func mapGroupChat(s *port.GroupChatSession) groupChatResponse {
	return groupChatResponse{
		APIKey:      s.APIKey,
		Token:       s.Token,
		ChannelType: s.ChannelType,
		ChannelID:   s.ChannelID,
		UserID:      s.UserID,
		UserName:    s.UserName,
	}
}

func (h *Handler) GetGroupChat(w http.ResponseWriter, r *http.Request) {
	if h.chat == nil {
		respondError(w, http.StatusServiceUnavailable, "unavailable", "Chat is not configured")
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	sess, err := h.chat.GroupSession(r.Context(), actorID, id)
	if err != nil {
		if errors.Is(err, chat.ErrChatDisabled) {
			respondError(w, http.StatusServiceUnavailable, "unavailable", "Chat is not configured")
			return
		}
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, mapGroupChat(sess))
}
