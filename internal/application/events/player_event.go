package events

import (
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// PlayerEventService manages RSVP / join for tee times.
type PlayerEventService struct {
	playerEvents port.PlayerEventRepository
	events       port.EventRepository
}

func NewPlayerEventService(playerEvents port.PlayerEventRepository, events port.EventRepository) *PlayerEventService {
	return &PlayerEventService{playerEvents: playerEvents, events: events}
}
