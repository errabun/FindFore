package courses

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) List(ctx context.Context) ([]entity.Course, error) {
	return s.courses.List(ctx)
}
