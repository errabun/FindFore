package feed

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) List(ctx context.Context, limit, offset int32) ([]entity.PostWithDetails, error) {
	return s.posts.List(ctx, limit, offset)
}
