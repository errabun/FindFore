package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
)

type friendshipRequest struct {
	FollowerID int32 `json:"follower_id"`
	FolloweeID int32 `json:"followee_id"`
}

func (h *Handler) CreateFriendship(w http.ResponseWriter, r *http.Request) {
	var req friendshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	followerID := req.FollowerID
	if authPlayerID, ok := r.Context().Value(middleware.PlayerIDKey).(int64); ok && authPlayerID > 0 {
		followerID = int32(authPlayerID)
	}

	if followerID <= 0 || req.FolloweeID <= 0 {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid follower or followee id")
		return
	}

	if followerID == req.FolloweeID {
		respondError(w, http.StatusBadRequest, "bad_request", "Cannot create friendship with yourself")
		return
	}

	friendship, followerDetails, followeeDetails, err := h.friendships.FindOrCreate(r.Context(), followerID, req.FolloweeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create friendship")
		return
	}

	resp := FriendshipResponse{
		ID:         friendship.ID,
		FollowerID: friendship.FollowerID,
		FolloweeID: friendship.FolloweeID,
		Follower:   mapPlayerToResponse(*followerDetails),
		Followee:   mapPlayerToResponse(*followeeDetails),
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) DeleteFriendship(w http.ResponseWriter, r *http.Request) {
	var req friendshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	followerID := req.FollowerID
	if authPlayerID, ok := r.Context().Value(middleware.PlayerIDKey).(int64); ok && authPlayerID > 0 {
		followerID = int32(authPlayerID)
	}

	if followerID <= 0 || req.FolloweeID <= 0 {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid follower or followee id")
		return
	}

	err := h.friendships.Delete(r.Context(), followerID, req.FolloweeID)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Friendship not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}
