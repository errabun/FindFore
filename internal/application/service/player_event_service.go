package service

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type PlayerEventService struct {
	playerEvents port.PlayerEventRepository
	events       port.EventRepository
}

func NewPlayerEventService(playerEvents port.PlayerEventRepository, events port.EventRepository) *PlayerEventService {
	return &PlayerEventService{playerEvents: playerEvents, events: events}
}

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

	pe, err := s.playerEvents.UpdateStatus(ctx, playerID, eventID, inviteStatus)
	if err != nil {
		return nil, fmt.Errorf("update player event status: %w", err)
	}

	if err := s.closeOrOpenInvitations(ctx, eventID); err != nil {
		return nil, fmt.Errorf("close or open invitations: %w", err)
	}

	return pe, nil
}

func (s *PlayerEventService) JoinEvent(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	// Check that the player isn't already part of this event
	_, err := s.playerEvents.Get(ctx, playerID, eventID)
	if err == nil {
		return nil, fmt.Errorf("player is already part of this event")
	}

	// Check remaining spots
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

	// Create player_event with accepted status
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
