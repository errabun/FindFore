package players

import (
	"regexp"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@(?:[-a-z0-9]+\.)+[a-z]{2,}$`)

// Service manages player profiles and credentials.
type Service struct {
	players     port.PlayerRepository
	friendships port.FriendshipRepository
}

func NewService(players port.PlayerRepository, friendships port.FriendshipRepository) *Service {
	return &Service{players: players, friendships: friendships}
}
