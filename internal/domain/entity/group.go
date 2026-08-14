package entity

import (
	"errors"
	"time"
)

const (
	GroupPrivacyPublic  = "public"
	GroupPrivacyPrivate = "private"

	GroupRoleOwner  = "owner"
	GroupRoleAdmin  = "admin"
	GroupRoleMember = "member"

	GroupMembershipActive  = "active"
	GroupMembershipPending = "pending"
)

var (
	ErrGroupNotFound        = errors.New("group not found")
	ErrGroupForbidden       = errors.New("group action forbidden for this player")
	ErrGroupConflict        = errors.New("group relationship conflict")
	ErrGroupOwnerCannotLeave = errors.New("group owner cannot leave")
)

// Group is a persistent collection of golfers.
type Group struct {
	ID            int64
	OwnerPlayerID int64
	Name          string
	Description   string
	Privacy       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (g Group) IsPublic() bool {
	return g.Privacy == GroupPrivacyPublic
}

// GroupMembership is a player's relationship to a group (request or active seat).
type GroupMembership struct {
	GroupID   int64
	PlayerID  int64
	Role      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m GroupMembership) IsActive() bool {
	return m.Status == GroupMembershipActive
}

func (m GroupMembership) IsOwner() bool {
	return m.Role == GroupRoleOwner && m.IsActive()
}

func (m GroupMembership) CanManage() bool {
	return m.IsActive() && (m.Role == GroupRoleOwner || m.Role == GroupRoleAdmin)
}

// GroupInvitation is an outstanding or terminal invite (not a membership row).
type GroupInvitation struct {
	ID                int64
	GroupID           int64
	InviterPlayerID   int64
	InviteePlayerID   int64
	CreatedAt         time.Time
	ExpiresAt         *time.Time
	AcceptedAt        *time.Time
	DeclinedAt        *time.Time
}

func (inv GroupInvitation) IsOutstanding(now time.Time) bool {
	if inv.AcceptedAt != nil || inv.DeclinedAt != nil {
		return false
	}
	if inv.ExpiresAt != nil && !inv.ExpiresAt.After(now) {
		return false
	}
	return true
}
