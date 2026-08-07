package friends

import (
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

var (
	ErrFriendshipNotFound       = errors.New("friendship not found")
	ErrFriendshipForbidden      = errors.New("not allowed to modify this friendship")
	ErrFriendshipSelf           = errors.New("cannot create friendship with yourself")
	ErrFriendshipAlreadyFriends = errors.New("already friends")
	ErrFriendshipAlreadyPending = errors.New("friend request already pending")
)

// Service manages friend requests and accepted friendships.
type Service struct {
	friendships port.FriendshipRepository
	playerSvc   port.PlayerService
}

func NewService(friendships port.FriendshipRepository, playerSvc port.PlayerService) *Service {
	return &Service{friendships: friendships, playerSvc: playerSvc}
}
