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
	courses      port.CourseRepository
	groups       port.GroupRepository
}

func NewService(events port.EventRepository, playerEvents port.PlayerEventRepository, courses port.CourseRepository, groups port.GroupRepository) *Service {
	return &Service{events: events, playerEvents: playerEvents, courses: courses, groups: groups}
}
