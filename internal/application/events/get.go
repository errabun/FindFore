package events

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Get(ctx context.Context, id, viewerID int64) (*entity.EventWithDetails, error) {
	details, err := s.buildDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.canView(details, viewerID) {
		return nil, ErrEventNotFound
	}
	return details, nil
}
