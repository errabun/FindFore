package courses

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) FindOrCreate(ctx context.Context, c entity.Course) (*entity.Course, error) {
	existing, err := s.courses.GetByNameAndCity(ctx, c.Name, c.City)
	if err == nil {
		return existing, nil
	}

	created, err := s.courses.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("create course: %w", err)
	}
	return created, nil
}
