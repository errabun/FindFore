package events

import (
	"context"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

// List returns events for the authenticated actor.
// When forPlayerID is set it must equal actorID (own invite/commitment list).
// Otherwise only public events are returned (never a dump of all private rounds).
func (s *Service) List(ctx context.Context, actorID int64, forPlayerID *int64, publicOnly bool) ([]entity.EventWithDetails, error) {
	today := time.Now().Format("2006-01-02")
	_ = s.events.DeletePast(ctx, today)

	var eventIDs []int64
	var err error

	if forPlayerID != nil {
		if *forPlayerID != actorID {
			return nil, ErrEventForbidden
		}
		eventIDs, err = s.events.ListIDsByPlayerID(ctx, actorID)
	} else {
		_ = publicOnly
		eventIDs, err = s.events.ListPublicIDs(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("list event IDs: %w", err)
	}

	result := make([]entity.EventWithDetails, 0, len(eventIDs))
	for _, eid := range eventIDs {
		details, err := s.buildDetails(ctx, eid)
		if err != nil {
			return nil, err
		}
		result = append(result, *details)
	}
	return result, nil
}

func (s *Service) ListFriendsEvents(ctx context.Context, actorID int64) ([]entity.EventWithDetails, error) {
	today := time.Now().Format("2006-01-02")
	_ = s.events.DeletePast(ctx, today)

	eventIDs, err := s.events.ListFriendsAvailableIDs(ctx, int32(actorID), actorID)
	if err != nil {
		return nil, fmt.Errorf("list friends available event IDs: %w", err)
	}

	result := make([]entity.EventWithDetails, 0, len(eventIDs))
	for _, eid := range eventIDs {
		details, err := s.buildDetails(ctx, eid)
		if err != nil {
			return nil, err
		}
		result = append(result, *details)
	}
	return result, nil
}
