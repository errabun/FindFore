package sessions

import (
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// Service handles authentication sessions (login / JWT issuance).
type Service struct {
	players     port.PlayerRepository
	friendships port.FriendshipRepository
	jwtSecret   string
}

func NewService(players port.PlayerRepository, friendships port.FriendshipRepository, jwtSecret string) *Service {
	return &Service{
		players:     players,
		friendships: friendships,
		jwtSecret:   jwtSecret,
	}
}
