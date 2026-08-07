package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/application/friends"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type createFriendshipRequest struct {
	PlayerID int32 `json:"player_id"`
}

func actorIDFromRequest(r *http.Request) (int32, bool) {
	authPlayerID, ok := middleware.PlayerIDFromContext(r.Context())
	if !ok {
		return 0, false
	}
	return int32(authPlayerID), true
}

func mapFriendshipResponse(
	f *entity.Friendship,
	requester, addressee *entity.PlayerWithDetails,
) FriendshipResponse {
	return FriendshipResponse{
		ID:          f.ID,
		RequesterID: f.RequesterID,
		AddresseeID: f.AddresseeID,
		Status:      f.Status.String(),
		Requester:   mapPlayerToResponse(*requester),
		Addressee:   mapPlayerToResponse(*addressee),
	}
}

func friendshipErrorStatus(err error) (int, string, string) {
	switch {
	case errors.Is(err, friends.ErrFriendshipNotFound):
		return http.StatusNotFound, "not_found", "Friendship not found"
	case errors.Is(err, friends.ErrFriendshipForbidden):
		return http.StatusForbidden, "forbidden", "Not allowed to modify this friendship"
	case errors.Is(err, friends.ErrFriendshipSelf):
		return http.StatusBadRequest, "bad_request", "Cannot create friendship with yourself"
	case errors.Is(err, friends.ErrFriendshipAlreadyFriends):
		return http.StatusConflict, "conflict", "Already friends"
	case errors.Is(err, friends.ErrFriendshipAlreadyPending):
		return http.StatusConflict, "conflict", "Friend request already pending"
	default:
		return http.StatusInternalServerError, "internal_error", "Friendship operation failed"
	}
}

func (h *Handler) CreateFriendship(w http.ResponseWriter, r *http.Request) {
	actorID, ok := actorIDFromRequest(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req createFriendshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}
	if req.PlayerID <= 0 {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid player_id")
		return
	}

	f, requester, addressee, err := h.friendships.Request(r.Context(), actorID, req.PlayerID)
	if err != nil {
		status, code, msg := friendshipErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	respondJSON(w, http.StatusCreated, mapFriendshipResponse(f, requester, addressee))
}

func (h *Handler) ListFriendships(w http.ResponseWriter, r *http.Request) {
	actorID, ok := actorIDFromRequest(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	rows, err := h.friendships.ListAccepted(r.Context(), actorID)
	if err != nil {
		respondInternalError(w, r, err, "Failed to list friendships")
		return
	}

	resp := make([]FriendshipResponse, 0, len(rows))
	for i := range rows {
		f := &rows[i]
		requester, err := h.players.GetWithDetails(r.Context(), int64(f.RequesterID))
		if err != nil {
			respondInternalError(w, r, err, "Failed to load friendship")
			return
		}
		addressee, err := h.players.GetWithDetails(r.Context(), int64(f.AddresseeID))
		if err != nil {
			respondInternalError(w, r, err, "Failed to load friendship")
			return
		}
		resp = append(resp, mapFriendshipResponse(f, requester, addressee))
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListFriendshipRequests(w http.ResponseWriter, r *http.Request) {
	actorID, ok := actorIDFromRequest(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	rows, err := h.friendships.ListIncomingRequests(r.Context(), actorID)
	if err != nil {
		respondInternalError(w, r, err, "Failed to list friend requests")
		return
	}

	resp := make([]FriendshipResponse, 0, len(rows))
	for i := range rows {
		f := &rows[i]
		requester, err := h.players.GetWithDetails(r.Context(), int64(f.RequesterID))
		if err != nil {
			respondInternalError(w, r, err, "Failed to load friend request")
			return
		}
		addressee, err := h.players.GetWithDetails(r.Context(), int64(f.AddresseeID))
		if err != nil {
			respondInternalError(w, r, err, "Failed to load friend request")
			return
		}
		resp = append(resp, mapFriendshipResponse(f, requester, addressee))
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) ListOutgoingFriendshipRequests(w http.ResponseWriter, r *http.Request) {
	actorID, ok := actorIDFromRequest(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	ids, err := h.friendships.ListOutgoingPendingIDs(r.Context(), actorID)
	if err != nil {
		respondInternalError(w, r, err, "Failed to list outgoing requests")
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	respondJSON(w, http.StatusOK, ids)
}

func (h *Handler) AcceptFriendship(w http.ResponseWriter, r *http.Request) {
	actorID, ok := actorIDFromRequest(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid friendship id")
		return
	}

	f, requester, addressee, err := h.friendships.Accept(r.Context(), actorID, id)
	if err != nil {
		status, code, msg := friendshipErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	respondJSON(w, http.StatusOK, mapFriendshipResponse(f, requester, addressee))
}

func (h *Handler) DeclineFriendship(w http.ResponseWriter, r *http.Request) {
	actorID, ok := actorIDFromRequest(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid friendship id")
		return
	}

	if err := h.friendships.Decline(r.Context(), actorID, id); err != nil {
		status, code, msg := friendshipErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	respondJSON(w, http.StatusOK, nil)
}

func (h *Handler) DeleteFriendship(w http.ResponseWriter, r *http.Request) {
	actorID, ok := actorIDFromRequest(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid friendship id")
		return
	}

	if err := h.friendships.CancelOrUnfriend(r.Context(), actorID, id); err != nil {
		status, code, msg := friendshipErrorStatus(err)
		respondLoggedError(w, r, status, code, msg, err)
		return
	}

	respondJSON(w, http.StatusOK, nil)
}
