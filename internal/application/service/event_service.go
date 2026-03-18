package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type EventService struct {
	events       port.EventRepository
	playerEvents port.PlayerEventRepository
}

func NewEventService(events port.EventRepository, playerEvents port.PlayerEventRepository) *EventService {
	return &EventService{events: events, playerEvents: playerEvents}
}

func (s *EventService) buildDetails(ctx context.Context, eventID int64) (*entity.EventWithDetails, error) {
	details, err := s.events.GetDetailsByID(ctx, eventID)
	if err != nil {
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

	return details, nil
}

func (s *EventService) List(ctx context.Context, playerID *int64, publicOnly bool) ([]entity.EventWithDetails, error) {
	today := time.Now().Format("2006-01-02")
	_ = s.events.DeletePast(ctx, today)

	var eventIDs []int64
	var err error

	if playerID != nil {
		eventIDs, err = s.events.ListIDsByPlayerID(ctx, *playerID)
	} else if publicOnly {
		eventIDs, err = s.events.ListPublicIDs(ctx)
	} else {
		eventIDs, err = s.events.ListAllIDs(ctx)
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

func (s *EventService) Get(ctx context.Context, id int64) (*entity.EventWithDetails, error) {
	return s.buildDetails(ctx, id)
}

func (s *EventService) Create(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error) {
	eventID, err := s.events.CreateWithInvites(ctx, e, invitees)
	if err != nil {
		return nil, fmt.Errorf("create event with invites: %w", err)
	}
	return s.buildDetails(ctx, eventID)
}

func (s *EventService) Update(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error) {
	if err := s.events.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("update event %d: %w", e.ID, err)
	}

	// Add new invitees (skip if they already have a player_event record)
	for _, inviteeID := range invitees {
		_, err := s.playerEvents.Get(ctx, inviteeID, e.ID)
		if err != nil {
			// No existing record — create a pending invite
			_, _ = s.playerEvents.Create(ctx, entity.PlayerEvent{
				PlayerID:     inviteeID,
				EventID:      e.ID,
				InviteStatus: entity.InviteStatusPending,
			})
		}
	}

	return s.buildDetails(ctx, e.ID)
}

func (s *EventService) Delete(ctx context.Context, id int64) error {
	return s.events.Delete(ctx, id)
}

func (s *EventService) ListFriendsEvents(ctx context.Context, playerID int64) ([]entity.EventWithDetails, error) {
	today := time.Now().Format("2006-01-02")
	_ = s.events.DeletePast(ctx, today)

	eventIDs, err := s.events.ListFriendsAvailableIDs(ctx, int32(playerID), playerID)
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
