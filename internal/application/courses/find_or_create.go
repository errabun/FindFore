package courses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

// FindOrCreate resolves a canonical course, optionally linking a provider identity.
// link may be nil when the client has no provider external id.
func (s *Service) FindOrCreate(ctx context.Context, c entity.Course, link *entity.CourseProvider) (*entity.Course, bool, error) {
	if link != nil && link.Provider != "" && link.ExternalID != "" {
		existing, err := s.courses.GetByProviderExternalID(ctx, link.Provider, link.ExternalID)
		if err == nil {
			return existing, false, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, false, fmt.Errorf("get course by provider: %w", err)
		}
	}

	existing, err := s.courses.GetByNameAndCity(ctx, c.Name, c.City)
	if err == nil {
		if err := s.linkProvider(ctx, existing.ID, link); err != nil {
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
	if err := s.linkProvider(ctx, created.ID, link); err != nil {
		return nil, false, err
	}
	return created, true, nil
}

func (s *Service) linkProvider(ctx context.Context, courseID int64, link *entity.CourseProvider) error {
	if link == nil || link.Provider == "" || link.ExternalID == "" {
		return nil
	}
	if err := s.courses.LinkProvider(ctx, courseID, link.Provider, link.ExternalID); err != nil {
		return fmt.Errorf("link course provider: %w", err)
	}
	return nil
}
