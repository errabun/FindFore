package groups

import (
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

const (
	maxNameLen        = 100
	maxDescriptionLen = 1000
	defaultListLimit  = 20
	maxListLimit      = 50
	invitationTTLDays = 30
)

var (
	ErrGroupNotFound         = errors.New("group not found")
	ErrGroupForbidden        = errors.New("group action forbidden for this player")
	ErrGroupConflict         = errors.New("group relationship conflict")
	ErrGroupOwnerCannotLeave = errors.New("group owner cannot leave")
	ErrInvalidGroup          = errors.New("invalid group")
	ErrInvitationNotFound    = errors.New("group invitation not found")
	ErrInvitationExpired     = errors.New("group invitation expired")
)

type Service struct {
	groups  port.GroupRepository
	players port.PlayerRepository
}

func NewService(groups port.GroupRepository, players port.PlayerRepository) *Service {
	return &Service{groups: groups, players: players}
}
