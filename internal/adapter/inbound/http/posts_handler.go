package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/application/feed"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func mapPostToResponse(p entity.PostWithDetails) PostResponse {
	reactions := make([]ReactionResponse, 0, len(p.Reactions))
	for _, rx := range p.Reactions {
		reactions = append(reactions, mapReactionToResponse(rx))
	}

	replies := make([]ReplyResponse, 0, len(p.Replies))
	for _, rp := range p.Replies {
		replies = append(replies, mapReplyToResponse(rp))
	}

	return PostResponse{
		ID:         p.ID,
		PlayerID:   p.PlayerID,
		PlayerName: p.PlayerName,
		Body:       p.Body,
		CreatedAt:  p.CreatedAt.Format(time.RFC3339),
		Reactions:  reactions,
		Replies:    replies,
	}
}

func mapReactionToResponse(rx entity.Reaction) ReactionResponse {
	return ReactionResponse{
		ID:         rx.ID,
		PlayerID:   rx.PlayerID,
		PlayerName: rx.PlayerName,
		Emoji:      rx.Emoji,
	}
}

func mapReplyToResponse(rp entity.Reply) ReplyResponse {
	return ReplyResponse{
		ID:         rp.ID,
		PlayerID:   rp.PlayerID,
		PlayerName: rp.PlayerName,
		Body:       rp.Body,
		CreatedAt:  rp.CreatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(50)
	offset := int32(0)

	if limitStr != "" {
		if v, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(v)
		}
	}
	if offsetStr != "" {
		if v, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int32(v)
		}
	}

	posts, err := h.posts.List(r.Context(), limit, offset)
	if err != nil {
		respondInternalError(w, r, err, "Failed to fetch posts")
		return
	}

	resp := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		resp = append(resp, mapPostToResponse(p))
	}

	respondJSON(w, http.StatusOK, resp)
}

func writePostError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	var ve *apperr.ValidationError
	switch {
	case errors.As(err, &ve):
		respondError(w, http.StatusBadRequest, "validation_error", ve.Message)
	case errors.Is(err, feed.ErrPostNotFound):
		respondError(w, http.StatusNotFound, "not_found", "Resource not found")
	default:
		respondInternalError(w, r, err, fallback)
	}
}

type createPostRequest struct {
	Body string `json:"body"`
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	post, err := h.posts.Create(r.Context(), actorID, req.Body)
	if err != nil {
		writePostError(w, r, err, "Failed to create post")
		return
	}

	respondJSON(w, http.StatusCreated, mapPostToResponse(*post))
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid post ID")
		return
	}

	// Body is optional for backwards compatibility; actor comes from JWT.
	_ = json.NewDecoder(r.Body).Decode(&struct{}{})

	if err := h.posts.Delete(r.Context(), postID, actorID); err != nil {
		writePostError(w, r, err, "Failed to delete post")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}

type toggleReactionRequest struct {
	Emoji string `json:"emoji"`
}

func (h *Handler) ToggleReaction(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid post ID")
		return
	}

	var req toggleReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	if req.Emoji == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Emoji is required")
		return
	}

	reactions, err := h.posts.ToggleReaction(r.Context(), postID, actorID, req.Emoji)
	if err != nil {
		writePostError(w, r, err, "Failed to toggle reaction")
		return
	}

	resp := make([]ReactionResponse, 0, len(reactions))
	for _, rx := range reactions {
		resp = append(resp, mapReactionToResponse(rx))
	}

	respondJSON(w, http.StatusOK, resp)
}

type createReplyRequest struct {
	Body string `json:"body"`
}

func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid post ID")
		return
	}

	var req createReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	if req.Body == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Reply body can't be blank")
		return
	}

	reply, err := h.posts.CreateReply(r.Context(), postID, actorID, req.Body)
	if err != nil {
		writePostError(w, r, err, "Failed to create reply")
		return
	}

	respondJSON(w, http.StatusCreated, mapReplyToResponse(*reply))
}

func (h *Handler) DeleteReply(w http.ResponseWriter, r *http.Request) {
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	replyIDStr := chi.URLParam(r, "reply_id")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid reply ID")
		return
	}

	_ = json.NewDecoder(r.Body).Decode(&struct{}{})

	if err := h.posts.DeleteReply(r.Context(), replyID, actorID); err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Reply not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}

func (h *Handler) requirePosts(w http.ResponseWriter) bool {
	if h.posts == nil {
		respondError(w, http.StatusNotImplemented, "not_implemented", "Feed is not configured")
		return false
	}
	return true
}

func (h *Handler) ListGroupPosts(w http.ResponseWriter, r *http.Request) {
	if !h.requirePosts(w) {
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
	limit, offset := parseLimitOffset(r)
	posts, err := h.posts.ListForGroup(r.Context(), actorID, id, limit, offset)
	if err != nil {
		writePostError(w, r, err, "Failed to fetch group posts")
		return
	}
	resp := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		resp = append(resp, mapPostToResponse(p))
	}
	respondJSON(w, http.StatusOK, map[string]any{"posts": resp})
}

func (h *Handler) CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	if !h.requirePosts(w) {
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
	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}
	post, err := h.posts.CreateForGroup(r.Context(), actorID, id, req.Body)
	if err != nil {
		writePostError(w, r, err, "Failed to create group post")
		return
	}
	respondJSON(w, http.StatusCreated, mapPostToResponse(*post))
}
