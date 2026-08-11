package courses

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Search(ctx context.Context, query string) ([]entity.CourseSearchResult, error) {
	return s.searcher.Search(ctx, query)
}
