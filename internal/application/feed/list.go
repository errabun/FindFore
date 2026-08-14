package feed

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) List(ctx context.Context, limit, offset int32) ([]entity.PostWithDetails, error) {
	return s.posts.List(ctx, clampLimit(limit), clampOffset(offset))
}

func (s *Service) ListForGroup(ctx context.Context, actorID, groupID int64, limit, offset int32) ([]entity.PostWithDetails, error) {
	if err := s.requireActiveMember(ctx, groupID, actorID); err != nil {
		return nil, err
	}
	return s.posts.ListByGroupID(ctx, groupID, clampLimit(limit), clampOffset(offset))
}
