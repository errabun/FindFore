package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type PostService struct {
	posts     port.PostRepository
	reactions port.ReactionRepository
	replies   port.ReplyRepository
}

func NewPostService(posts port.PostRepository, reactions port.ReactionRepository, replies port.ReplyRepository) *PostService {
	return &PostService{posts: posts, reactions: reactions, replies: replies}
}

func (s *PostService) List(ctx context.Context, limit, offset int32) ([]entity.PostWithDetails, error) {
	return s.posts.List(ctx, limit, offset)
}

func (s *PostService) Create(ctx context.Context, playerID int64, body string) (*entity.PostWithDetails, error) {
	if body == "" {
		return nil, &ValidationError{Message: "Post body can't be blank"}
	}

	id, err := s.posts.Create(ctx, playerID, body)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post %d: %w", id, err)
	}

	return post, nil
}

func (s *PostService) Delete(ctx context.Context, postID, playerID int64) error {
	return s.posts.Delete(ctx, postID, playerID)
}

func (s *PostService) ToggleReaction(ctx context.Context, postID, playerID int64, emoji string) ([]entity.Reaction, error) {
	_, err := s.reactions.Find(ctx, postID, playerID, emoji)
	if err == sql.ErrNoRows {
		// Doesn't exist, create it
		_, err = s.reactions.Create(ctx, postID, playerID, emoji)
		if err != nil {
			return nil, fmt.Errorf("create reaction: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("find reaction: %w", err)
	} else {
		// Exists, delete it
		if err := s.reactions.Delete(ctx, postID, playerID, emoji); err != nil {
			return nil, fmt.Errorf("delete reaction: %w", err)
		}
	}

	return s.reactions.ListByPostID(ctx, postID)
}

func (s *PostService) CreateReply(ctx context.Context, postID, playerID int64, body string) (*entity.Reply, error) {
	if body == "" {
		return nil, &ValidationError{Message: "Reply body can't be blank"}
	}

	reply, err := s.replies.Create(ctx, postID, playerID, body)
	if err != nil {
		return nil, fmt.Errorf("create reply: %w", err)
	}

	return reply, nil
}

func (s *PostService) DeleteReply(ctx context.Context, replyID, playerID int64) error {
	return s.replies.Delete(ctx, replyID, playerID)
}
