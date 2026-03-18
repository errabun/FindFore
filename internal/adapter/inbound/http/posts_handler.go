package httphandler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

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
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch posts")
		return
	}

	resp := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		resp = append(resp, mapPostToResponse(p))
	}

	respondJSON(w, http.StatusOK, resp)
}

type createPostRequest struct {
	PlayerID int64  `json:"player_id"`
	Body     string `json:"body"`
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	if req.PlayerID == 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "Player ID is required")
		return
	}

	post, err := h.posts.Create(r.Context(), req.PlayerID, req.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create post")
		return
	}

	respondJSON(w, http.StatusCreated, mapPostToResponse(*post))
}

type deletePostRequest struct {
	PlayerID int64 `json:"player_id"`
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid post ID")
		return
	}

	var req deletePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	if err := h.posts.Delete(r.Context(), postID, req.PlayerID); err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Post not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}

type toggleReactionRequest struct {
	PlayerID int64  `json:"player_id"`
	Emoji    string `json:"emoji"`
}

func (h *Handler) ToggleReaction(w http.ResponseWriter, r *http.Request) {
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

	reactions, err := h.posts.ToggleReaction(r.Context(), postID, req.PlayerID, req.Emoji)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to toggle reaction")
		return
	}

	resp := make([]ReactionResponse, 0, len(reactions))
	for _, rx := range reactions {
		resp = append(resp, mapReactionToResponse(rx))
	}

	respondJSON(w, http.StatusOK, resp)
}

type createReplyRequest struct {
	PlayerID int64  `json:"player_id"`
	Body     string `json:"body"`
}

func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request) {
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

	reply, err := h.posts.CreateReply(r.Context(), postID, req.PlayerID, req.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create reply")
		return
	}

	respondJSON(w, http.StatusCreated, mapReplyToResponse(*reply))
}

func (h *Handler) DeleteReply(w http.ResponseWriter, r *http.Request) {
	replyIDStr := chi.URLParam(r, "reply_id")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid reply ID")
		return
	}

	var req struct {
		PlayerID int64 `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "bad_request", "Invalid request body")
		return
	}

	if err := h.posts.DeleteReply(r.Context(), replyID, req.PlayerID); err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Reply not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}
