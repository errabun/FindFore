package courses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

// FindOrCreate returns the course and whether it was newly created.
func (s *Service) FindOrCreate(ctx context.Context, c entity.Course) (*entity.Course, bool, error) {
	if c.Provider != "" && c.ExternalID != "" {
		existing, err := s.courses.GetByProviderExternalID(ctx, c.Provider, c.ExternalID)
		if err == nil {
			return existing, false, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, false, fmt.Errorf("get course by provider: %w", err)
		}
	}

	existing, err := s.courses.GetByNameAndCity(ctx, c.Name, c.City)
	if err == nil {
		if err := s.linkProvider(ctx, existing.ID, c); err != nil {
			return nil, false, err
		}
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("get course by name/city: %w", err)
	}

	created, err := s.courses.Create(ctx, c)
	if err != nil {
		return nil, false, fmt.Errorf("create course: %w", err)
	}
	if err := s.linkProvider(ctx, created.ID, c); err != nil {
		return nil, false, err
	}
	return created, true, nil
}

func (s *Service) linkProvider(ctx context.Context, courseID int64, c entity.Course) error {
	if c.Provider == "" || c.ExternalID == "" {
		return nil
	}
	if err := s.courses.UpsertProvider(ctx, courseID, c.Provider, c.ExternalID); err != nil {
		return fmt.Errorf("upsert course provider: %w", err)
	}
	return nil
}
