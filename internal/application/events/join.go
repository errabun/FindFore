package events

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *PlayerEventService) JoinEvent(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	_, err := s.playerEvents.Get(ctx, playerID, eventID)
	if err == nil {
		return nil, fmt.Errorf("player is already part of this event")
	}

	event, err := s.events.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("get event %d: %w", eventID, err)
	}

	acceptedCount, err := s.playerEvents.CountAcceptedForEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("count accepted for event %d: %w", eventID, err)
	}

	remaining := event.OpenSpots - int32(acceptedCount)
	if remaining <= 0 {
		return nil, fmt.Errorf("event is full")
	}

	pe, err := s.playerEvents.Create(ctx, entity.PlayerEvent{
		PlayerID:     playerID,
		EventID:      eventID,
		InviteStatus: entity.InviteStatusAccepted,
	})
	if err != nil {
		return nil, fmt.Errorf("create player event: %w", err)
	}

	if err := s.closeOrOpenInvitations(ctx, eventID); err != nil {
		return nil, fmt.Errorf("close or open invitations: %w", err)
	}

	return pe, nil
}
