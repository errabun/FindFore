package events

import (
	"context"
	"fmt"
	"sort"

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
		s.attachGroupName(ctx, details)
		result = append(result, *details)
	}
	return result, nil
}

const maxJoinableGroups = 50

// ListJoinableFromGroups returns upcoming group rounds the actor can still join:
// active member of the group, not already accepted, remaining spots > 0.
func (s *Service) ListJoinableFromGroups(ctx context.Context, actorID int64) ([]entity.EventWithDetails, error) {
	if actorID <= 0 || s.groups == nil {
		return []entity.EventWithDetails{}, nil
	}
	_ = s.events.DeletePast(ctx)

	groups, err := s.groups.ListByPlayer(ctx, actorID, maxJoinableGroups, 0)
	if err != nil {
		return nil, fmt.Errorf("list actor groups: %w", err)
	}

	result := make([]entity.EventWithDetails, 0)
	seen := make(map[int64]bool, 8)
	for _, g := range groups {
		ids, err := s.events.ListIDsByGroupID(ctx, g.ID)
		if err != nil {
			return nil, fmt.Errorf("list events for group %d: %w", g.ID, err)
		}
		for _, eid := range ids {
			if seen[eid] {
				continue
			}
			seen[eid] = true
			details, err := s.buildDetails(ctx, eid)
			if err != nil {
				return nil, err
			}
			if containsPlayer(details.Accepted, actorID) || details.RemainingSpots <= 0 {
				continue
			}
			details.GroupName = g.Name
			result = append(result, *details)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].PlannedStartsAt.Equal(result[j].PlannedStartsAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].PlannedStartsAt.Before(result[j].PlannedStartsAt)
	})
	return result, nil
}

func (s *Service) attachGroupName(ctx context.Context, details *entity.EventWithDetails) {
	if details == nil || details.GroupID == nil || s.groups == nil {
		return
	}
	g, err := s.groups.GetByID(ctx, *details.GroupID)
	if err != nil {
		return
	}
	details.GroupName = g.Name
}
