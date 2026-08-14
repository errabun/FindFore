package events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) buildDetails(ctx context.Context, eventID int64) (*entity.EventWithDetails, error) {
	details, err := s.events.GetDetailsByID(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("get event details %d: %w", eventID, err)
	}

	accepted, err := s.playerEvents.ListPlayerIDsByEventAndStatus(ctx, eventID, entity.InviteStatusAccepted)
	if err != nil {
		return nil, fmt.Errorf("list accepted for event %d: %w", eventID, err)
	}

	declined, err := s.playerEvents.ListPlayerIDsByEventAndStatus(ctx, eventID, entity.InviteStatusDeclined)
	if err != nil {
		return nil, fmt.Errorf("list declined for event %d: %w", eventID, err)
	}

	pending, err := s.playerEvents.ListPlayerIDsByEventAndStatus(ctx, eventID, entity.InviteStatusPending)
	if err != nil {
		return nil, fmt.Errorf("list pending for event %d: %w", eventID, err)
	}

	closed, err := s.playerEvents.ListPlayerIDsByEventAndStatus(ctx, eventID, entity.InviteStatusClosed)
	if err != nil {
		return nil, fmt.Errorf("list closed for event %d: %w", eventID, err)
	}

	details.Accepted = accepted
	details.Declined = declined
	details.Pending = pending
	details.Closed = closed
	details.RemainingSpots = details.OpenSpots - int32(len(accepted))

	date, teeTime, err := SplitStartsAt(details.PlannedStartsAt, details.CourseTimezone)
	if err != nil {
		return nil, fmt.Errorf("split planned_starts_at for event %d: %w", eventID, err)
	}
	details.Date = date
	details.TeeTime = teeTime

	return details, nil
}

func containsPlayer(ids []int64, playerID int64) bool {
	for _, id := range ids {
		if id == playerID {
			return true
		}
	}
	return false
}

func (s *Service) canView(ctx context.Context, details *entity.EventWithDetails, viewerID int64) bool {
	if !details.Private {
		return true
	}
	if int64(details.HostID) == viewerID {
		return true
	}
	if containsPlayer(details.Accepted, viewerID) ||
		containsPlayer(details.Pending, viewerID) ||
		containsPlayer(details.Declined, viewerID) ||
		containsPlayer(details.Closed, viewerID) {
		return true
	}
	if details.GroupID != nil {
		return requireActiveGroupMember(ctx, s.groups, *details.GroupID, viewerID) == nil
	}
	return false
}
