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

func (s *Service) hydrateRSVPs(ctx context.Context, details []entity.EventWithDetails) error {
	if len(details) == 0 {
		return nil
	}
	ids := make([]int64, len(details))
	for i := range details {
		ids[i] = details[i].ID
	}
	rows, err := s.playerEvents.ListByEventIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("list player events: %w", err)
	}

	accepted := make(map[int64][]int64, len(details))
	declined := make(map[int64][]int64, len(details))
	pending := make(map[int64][]int64, len(details))
	closed := make(map[int64][]int64, len(details))
	for _, pe := range rows {
		switch pe.InviteStatus {
		case entity.InviteStatusAccepted:
			accepted[pe.EventID] = append(accepted[pe.EventID], pe.PlayerID)
		case entity.InviteStatusDeclined:
			declined[pe.EventID] = append(declined[pe.EventID], pe.PlayerID)
		case entity.InviteStatusPending:
			pending[pe.EventID] = append(pending[pe.EventID], pe.PlayerID)
		case entity.InviteStatusClosed:
			closed[pe.EventID] = append(closed[pe.EventID], pe.PlayerID)
		}
	}

	for i := range details {
		d := &details[i]
		d.Accepted = accepted[d.ID]
		d.Declined = declined[d.ID]
		d.Pending = pending[d.ID]
		d.Closed = closed[d.ID]
		d.RemainingSpots = d.OpenSpots - int32(len(d.Accepted))
		date, teeTime, err := SplitStartsAt(d.PlannedStartsAt, d.CourseTimezone)
		if err != nil {
			return fmt.Errorf("split planned_starts_at for event %d: %w", d.ID, err)
		}
		d.Date = date
		d.TeeTime = teeTime
	}
	return nil
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
