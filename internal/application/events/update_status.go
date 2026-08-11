package events

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *PlayerEventService) closeOrOpenInvitations(ctx context.Context, eventID int64) error {
	event, err := s.events.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("get event %d: %w", eventID, err)
	}

	acceptedCount, err := s.playerEvents.CountAcceptedForEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("count accepted for event %d: %w", eventID, err)
	}

	remaining := event.OpenSpots - int32(acceptedCount)

	if remaining <= 0 {
		return s.playerEvents.ClosePendingForEvent(ctx, eventID)
	}
	return s.playerEvents.ReopenClosedForEvent(ctx, eventID)
}

func (s *PlayerEventService) UpdateStatus(ctx context.Context, playerID, eventID int64, status string) (*entity.PlayerEvent, error) {
	inviteStatus, ok := entity.ParseInviteStatus(status)
	if !ok {
		return nil, fmt.Errorf("invalid invite status: %s", status)
	}

	if inviteStatus == entity.InviteStatusAccepted {
		return s.playerEvents.AcceptInvite(ctx, playerID, eventID)
	}

	pe, err := s.playerEvents.UpdateStatus(ctx, playerID, eventID, inviteStatus)
	if err != nil {
		return nil, fmt.Errorf("update player event status: %w", err)
	}

	if err := s.closeOrOpenInvitations(ctx, eventID); err != nil {
		return nil, fmt.Errorf("close or open invitations: %w", err)
	}

	return pe, nil
}
