package events

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

// List returns events for the authenticated actor.
// When forPlayerID is set it must equal actorID (own invite/commitment list).
// Otherwise only public events are returned (never a dump of all private rounds).
func (s *Service) List(ctx context.Context, actorID int64, forPlayerID *int64, publicOnly bool) ([]entity.EventWithDetails, error) {
	_ = s.events.DeletePast(ctx)

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
	_ = s.events.DeletePast(ctx)

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

func (s *Service) ListForGroup(ctx context.Context, actorID, groupID int64) ([]entity.EventWithDetails, error) {
	if err := requireActiveGroupMember(ctx, s.groups, groupID, actorID); err != nil {
		return nil, err
	}
	_ = s.events.DeletePast(ctx)

	eventIDs, err := s.events.ListIDsByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list group event IDs: %w", err)
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
