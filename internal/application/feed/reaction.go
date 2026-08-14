package feed

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) ToggleReaction(ctx context.Context, postID, playerID int64, emoji string) ([]entity.Reaction, error) {
	post, err := s.loadPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if err := s.requireCanRead(ctx, post, playerID); err != nil {
		return nil, err
	}

	_, err = s.reactions.Find(ctx, postID, playerID, emoji)
	if err == sql.ErrNoRows {
		_, err = s.reactions.Create(ctx, postID, playerID, emoji)
		if err != nil {
			return nil, fmt.Errorf("create reaction: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("find reaction: %w", err)
	} else {
		if err := s.reactions.Delete(ctx, postID, playerID, emoji); err != nil {
			return nil, fmt.Errorf("delete reaction: %w", err)
		}
	}

	return s.reactions.ListByPostID(ctx, postID)
}
