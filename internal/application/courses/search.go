package courses

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Search(ctx context.Context, query string) ([]entity.Course, error) {
	return s.searcher.Search(ctx, query)
}
