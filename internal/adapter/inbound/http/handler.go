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
	booking      port.BookingService
	groups       port.GroupService
	chat         port.ChatService
}

func New(
	players port.PlayerService,
	sessions port.SessionService,
	courses port.CourseService,
	events port.EventService,
	playerEvents port.PlayerEventService,
	friendships port.FriendshipService,
	posts port.PostService,
	bookingSvc port.BookingService,
	groupsSvc port.GroupService,
) *Handler {
	return &Handler{
		players:      players,
		sessions:     sessions,
		courses:      courses,
		events:       events,
		playerEvents: playerEvents,
		friendships:  friendships,
		posts:        posts,
		booking:      bookingSvc,
		groups:       groupsSvc,
	}
}

func (h *Handler) WithChat(chat port.ChatService) *Handler {
	h.chat = chat
	return h
}
