package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

var (
	ErrEventNotFound  = errors.New("event not found")
	ErrEventForbidden = errors.New("not allowed to access this event")
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

func (s *EventService) canView(details *entity.EventWithDetails, viewerID int64) bool {
	if !details.Private {
		return true
	}
	if int64(details.HostID) == viewerID {
		return true
	}
	return containsPlayer(details.Accepted, viewerID) ||
		containsPlayer(details.Pending, viewerID) ||
		containsPlayer(details.Declined, viewerID) ||
		containsPlayer(details.Closed, viewerID)
}

// List returns events for the authenticated actor.
// When forPlayerID is set it must equal actorID (own invite/commitment list).
// Otherwise only public events are returned (never a dump of all private rounds).
func (s *EventService) List(ctx context.Context, actorID int64, forPlayerID *int64, publicOnly bool) ([]entity.EventWithDetails, error) {
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
		// Default / ?private=false → public feed only
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

func (s *EventService) Get(ctx context.Context, id, viewerID int64) (*entity.EventWithDetails, error) {
	details, err := s.buildDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.canView(details, viewerID) {
		return nil, ErrEventNotFound
	}
	return details, nil
}

func (s *EventService) Create(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error) {
	eventID, err := s.events.CreateWithInvites(ctx, e, invitees)
	if err != nil {
		return nil, fmt.Errorf("create event with invites: %w", err)
	}
	return s.buildDetails(ctx, eventID)
}

func (s *EventService) Update(ctx context.Context, actorID int64, e entity.Event, invitees []int64) (*entity.EventWithDetails, error) {
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

	// Preserve host; never allow host reassignment via update body.
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

func (s *EventService) Delete(ctx context.Context, actorID, id int64) error {
	existing, err := s.events.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEventNotFound
		}
		return fmt.Errorf("get event %d: %w", id, err)
	}
	if int64(existing.HostID) != actorID {
		return ErrEventForbidden
	}
	return s.events.Delete(ctx, id)
}

func (s *EventService) ListFriendsEvents(ctx context.Context, actorID int64) ([]entity.EventWithDetails, error) {
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
