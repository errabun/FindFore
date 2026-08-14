package feed

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Create(ctx context.Context, playerID int64, body string) (*entity.PostWithDetails, error) {
	return s.create(ctx, playerID, body, nil)
}

func (s *Service) CreateForGroup(ctx context.Context, actorID, groupID int64, body string) (*entity.PostWithDetails, error) {
	if err := s.requireActiveMember(ctx, groupID, actorID); err != nil {
		return nil, err
	}
	return s.create(ctx, actorID, body, &groupID)
}

func (s *Service) create(ctx context.Context, playerID int64, body string, groupID *int64) (*entity.PostWithDetails, error) {
	if body == "" {
		return nil, &apperr.ValidationError{Message: "Post body can't be blank"}
	}

	id, err := s.posts.Create(ctx, playerID, body, groupID)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post %d: %w", id, err)
	}

	return post, nil
}
