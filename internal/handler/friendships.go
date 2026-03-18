package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ericrabun/findfore-go/internal/middleware"
	"github.com/ericrabun/findfore-go/internal/model"
	"github.com/ericrabun/findfore-go/internal/store"
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

	// Find-or-create pattern
	existing, err := h.queries.FindFriendship(r.Context(), store.FindFriendshipParams{
		FollowerID: sql.NullInt32{Int32: followerID, Valid: true},
		FolloweeID: sql.NullInt32{Int32: req.FolloweeID, Valid: true},
	})

	var friendshipID int64
	var friendshipFollowerID, followeeID int32

	if err == nil {
		// Already exists
		friendshipID = existing.ID
		friendshipFollowerID = existing.FollowerID.Int32
		followeeID = existing.FolloweeID.Int32
	} else {
		// Create new
		friendship, err := h.queries.CreateFriendship(r.Context(), store.CreateFriendshipParams{
			FollowerID: sql.NullInt32{Int32: followerID, Valid: true},
			FolloweeID: sql.NullInt32{Int32: req.FolloweeID, Valid: true},
		})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create friendship")
			return
		}
		friendshipID = friendship.ID
		friendshipFollowerID = friendship.FollowerID.Int32
		followeeID = friendship.FolloweeID.Int32
	}

	// Get full player details for follower and followee
	followerDetails, err := store.GetPlayerWithDetails(r.Context(), h.queries, int64(friendshipFollowerID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch follower details")
		return
	}

	followeeDetails, err := store.GetPlayerWithDetails(r.Context(), h.queries, int64(followeeID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch followee details")
		return
	}

	resp := model.FriendshipResponse{
		ID:         friendshipID,
		FollowerID: friendshipFollowerID,
		FolloweeID: followeeID,
		Follower: model.PlayerResponse{
			ID:       followerDetails.ID,
			Name:     followerDetails.Name,
			Phone:    followerDetails.Phone,
			Email:    followerDetails.Email,
			Username: followerDetails.Username,
			Friends:  followerDetails.Friends,
			Events:   followerDetails.Events,
		},
		Followee: model.PlayerResponse{
			ID:       followeeDetails.ID,
			Name:     followeeDetails.Name,
			Phone:    followeeDetails.Phone,
			Email:    followeeDetails.Email,
			Username: followeeDetails.Username,
			Friends:  followeeDetails.Friends,
			Events:   followeeDetails.Events,
		},
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

	err := h.queries.DeleteFriendship(r.Context(), store.DeleteFriendshipParams{
		FollowerID: sql.NullInt32{Int32: followerID, Valid: true},
		FolloweeID: sql.NullInt32{Int32: req.FolloweeID, Valid: true},
	})
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Friendship not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}
