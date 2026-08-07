package events

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Create(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error) {
	eventID, err := s.events.CreateWithInvites(ctx, e, invitees)
	if err != nil {
		return nil, fmt.Errorf("create event with invites: %w", err)
	}
	return s.buildDetails(ctx, eventID)
}
