package feed

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) CreateReply(ctx context.Context, postID, playerID int64, body string) (*entity.Reply, error) {
	if body == "" {
		return nil, &apperr.ValidationError{Message: "Reply body can't be blank"}
	}
	post, err := s.loadPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if err := s.requireCanRead(ctx, post, playerID); err != nil {
		return nil, err
	}

	reply, err := s.replies.Create(ctx, postID, playerID, body)
	if err != nil {
		return nil, fmt.Errorf("create reply: %w", err)
	}

	return reply, nil
}

func (s *Service) DeleteReply(ctx context.Context, replyID, playerID int64) error {
	return s.replies.Delete(ctx, replyID, playerID)
}
