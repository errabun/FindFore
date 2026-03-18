package httphandler

import (
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type Handler struct {
	players      port.PlayerService
	sessions     port.SessionService
	courses      port.CourseService
	events       port.EventService
	playerEvents port.PlayerEventService
	friendships  port.FriendshipService
	posts        port.PostService
}

func New(
	players port.PlayerService,
	sessions port.SessionService,
	courses port.CourseService,
	events port.EventService,
	playerEvents port.PlayerEventService,
	friendships port.FriendshipService,
	posts port.PostService,
) *Handler {
	return &Handler{
		players:      players,
		sessions:     sessions,
		courses:      courses,
		events:       events,
		playerEvents: playerEvents,
		friendships:  friendships,
		posts:        posts,
	}
}
