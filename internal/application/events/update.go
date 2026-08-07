package events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Update(ctx context.Context, actorID int64, e entity.Event, invitees []int64) (*entity.EventWithDetails, error) {
	existing, err := s.events.GetByID(ctx, e.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("get event %d: %w", e.ID, err)
	}
	if int64(existing.HostID) != actorID {
		return nil, ErrEventForbidden
	}

	e.HostID = existing.HostID

	if err := s.events.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update event %d: %w", e.ID, err)
	}

	for _, inviteeID := range invitees {
		_, err := s.playerEvents.Get(ctx, inviteeID, e.ID)
		if err != nil {
			_, _ = s.playerEvents.Create(ctx, entity.PlayerEvent{
				PlayerID:     inviteeID,
				EventID:      e.ID,
				InviteStatus: entity.InviteStatusPending,
			})
		}
	}

	return s.buildDetails(ctx, e.ID)
}
