package events

import (
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

var (
	ErrEventNotFound  = errors.New("event not found")
	ErrEventForbidden = errors.New("not allowed to access this event")
)

// Service manages tee time events (create, invite visibility, host updates).
type Service struct {
	events       port.EventRepository
	playerEvents port.PlayerEventRepository
}

func NewService(events port.EventRepository, playerEvents port.PlayerEventRepository) *Service {
	return &Service{events: events, playerEvents: playerEvents}
}
