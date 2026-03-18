package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ericrabun/findfore-go/internal/model"
	"github.com/ericrabun/findfore-go/internal/store"
)

// --- helpers ---

func (h *Handler) buildPostResponse(r *http.Request, postRow store.GetPostByIDRow) (*model.PostResponse, error) {
	reactions, err := h.queries.ListReactionsByPostID(r.Context(), postRow.ID)
	if err != nil {
		return nil, err
	}

	reactionResps := make([]model.ReactionResponse, 0, len(reactions))
	for _, rx := range reactions {
		reactionResps = append(reactionResps, model.ReactionResponse{
			ID:         rx.ID,
			PlayerID:   rx.PlayerID,
			PlayerName: rx.PlayerName.String,
			Emoji:      rx.Emoji,
		})
	}

	replies, err := h.queries.ListRepliesByPostID(r.Context(), postRow.ID)
	if err != nil {
		return nil, err
	}

	replyResps := make([]model.ReplyResponse, 0, len(replies))
	for _, rp := range replies {
		replyResps = append(replyResps, model.ReplyResponse{
			ID:         rp.ID,
			PlayerID:   rp.PlayerID,
			PlayerName: rp.PlayerName.String,
			Body:       rp.Body,
			CreatedAt:  rp.CreatedAt.Format(time.RFC3339),
		})
	}

	return &model.PostResponse{
		ID:         postRow.ID,
		PlayerID:   postRow.PlayerID,
		PlayerName: postRow.PlayerName.String,
		Body:       postRow.Body,
		CreatedAt:  postRow.CreatedAt.Format(time.RFC3339),
		Reactions:  reactionResps,
		Replies:    replyResps,
	}, nil
}

// --- Posts ---

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

	if req.Body == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Post body can't be blank")
		return
	}
	if req.PlayerID == 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "Player ID is required")
		return
	}

	created, err := h.queries.CreatePost(r.Context(), store.CreatePostParams{
		PlayerID: req.PlayerID,
		Body:     req.Body,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create post")
		return
	}

	post, err := h.queries.GetPostByID(r.Context(), created.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch post")
		return
	}

	resp, err := h.buildPostResponse(r, post)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to build post response")
		return
	}

	respondJSON(w, http.StatusCreated, resp)
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

	rows, err := h.queries.ListPosts(r.Context(), store.ListPostsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch posts")
		return
	}

	posts := make([]model.PostResponse, 0, len(rows))
	for _, row := range rows {
		postRow := store.GetPostByIDRow{
			ID:         row.ID,
			PlayerID:   row.PlayerID,
			Body:       row.Body,
			CreatedAt:  row.CreatedAt,
			PlayerName: row.PlayerName,
		}
		resp, err := h.buildPostResponse(r, postRow)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to build post response")
			return
		}
		posts = append(posts, *resp)
	}

	respondJSON(w, http.StatusOK, posts)
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

	if err := h.queries.DeletePost(r.Context(), store.DeletePostParams{
		ID:       postID,
		PlayerID: req.PlayerID,
	}); err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Post not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}

// --- Reactions ---

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

	// Toggle: if exists, delete; if not, create
	_, err = h.queries.FindReaction(r.Context(), store.FindReactionParams{
		PostID:   postID,
		PlayerID: req.PlayerID,
		Emoji:    req.Emoji,
	})
	if err == sql.ErrNoRows {
		// Doesn't exist, create it
		_, err = h.queries.CreateReaction(r.Context(), store.CreateReactionParams{
			PostID:   postID,
			PlayerID: req.PlayerID,
			Emoji:    req.Emoji,
		})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to add reaction")
			return
		}
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to check reaction")
		return
	} else {
		// Exists, delete it
		if err := h.queries.DeleteReaction(r.Context(), store.DeleteReactionParams{
			PostID:   postID,
			PlayerID: req.PlayerID,
			Emoji:    req.Emoji,
		}); err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", "Failed to remove reaction")
			return
		}
	}

	// Return updated reactions for the post
	reactions, err := h.queries.ListReactionsByPostID(r.Context(), postID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch reactions")
		return
	}

	resps := make([]model.ReactionResponse, 0, len(reactions))
	for _, rx := range reactions {
		resps = append(resps, model.ReactionResponse{
			ID:         rx.ID,
			PlayerID:   rx.PlayerID,
			PlayerName: rx.PlayerName.String,
			Emoji:      rx.Emoji,
		})
	}

	respondJSON(w, http.StatusOK, resps)
}

// --- Replies ---

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

	created, err := h.queries.CreateReply(r.Context(), store.CreateReplyParams{
		PostID:   postID,
		PlayerID: req.PlayerID,
		Body:     req.Body,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create reply")
		return
	}

	reply, err := h.queries.GetReplyByID(r.Context(), created.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch reply")
		return
	}

	resp := model.ReplyResponse{
		ID:         reply.ID,
		PlayerID:   reply.PlayerID,
		PlayerName: reply.PlayerName.String,
		Body:       reply.Body,
		CreatedAt:  reply.CreatedAt.Format(time.RFC3339),
	}

	respondJSON(w, http.StatusCreated, resp)
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

	if err := h.queries.DeleteReply(r.Context(), store.DeleteReplyParams{
		ID:       replyID,
		PlayerID: req.PlayerID,
	}); err != nil {
		respondError(w, http.StatusNotFound, "not_found", "Reply not found")
		return
	}

	respondJSON(w, http.StatusOK, nil)
}
