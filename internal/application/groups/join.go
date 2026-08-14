package groups

import (
	"context"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func (s *Service) Join(ctx context.Context, actorID, groupID int64) (*entity.GroupMembership, error) {
	if err := validateIDs(actorID, groupID); err != nil {
		return nil, err
	}
	g, err := s.loadGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	existing, err := s.membership(ctx, groupID, actorID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // idempotent join / pending request
	}

	if inv, err := s.outstandingInvite(ctx, groupID, actorID); err != nil {
		return nil, err
	} else if inv != nil {
		return s.groups.AcceptInvitation(ctx, inv.ID, actorID)
	}

	status := entity.GroupMembershipActive
	if !g.IsPublic() {
		status = entity.GroupMembershipPending
	}
	m, err := s.groups.InsertMembership(ctx, entity.GroupMembership{
		GroupID:  groupID,
		PlayerID: actorID,
		Role:     entity.GroupRoleMember,
		Status:   status,
	})
	if err != nil {
		if errors.Is(err, entity.ErrGroupConflict) {
			existing, getErr := s.membership(ctx, groupID, actorID)
			if getErr != nil {
				return nil, getErr
			}
			if existing != nil {
				return existing, nil
			}
			return nil, ErrGroupConflict
		}
		return nil, err
	}
	return m, nil
}

func (s *Service) Leave(ctx context.Context, actorID, groupID int64) error {
	if err := validateIDs(actorID, groupID); err != nil {
		return err
	}
	if _, err := s.loadGroup(ctx, groupID); err != nil {
		return err
	}
	m, err := s.membership(ctx, groupID, actorID)
	if err != nil {
		return err
	}
	if m == nil {
		return nil // already not a member
	}
	if m.IsOwner() {
		return ErrGroupOwnerCannotLeave
	}
	return s.groups.DeleteMembership(ctx, groupID, actorID)
}

func (s *Service) ListMembers(ctx context.Context, actorID, groupID int64, limit, offset int32) ([]port.GroupMember, error) {
	if err := validateIDs(actorID, groupID); err != nil {
		return nil, err
	}
	g, err := s.loadGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	ok, viewer, err := s.canView(ctx, g, actorID)
	if err != nil {
		return nil, err
	}
	if !ok || viewer == nil || !viewer.IsActive() {
		return nil, ErrGroupNotFound
	}
	rows, err := s.groups.ListActiveMembers(ctx, groupID, clampLimit(limit), clampOffset(offset))
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupMember, len(rows))
	for i, row := range rows {
		out[i] = port.GroupMember{
			PlayerID:   row.Membership.PlayerID,
			PlayerName: row.PlayerName,
			Role:       row.Membership.Role,
			Status:     row.Membership.Status,
		}
	}
	return out, nil
}

func (s *Service) RemoveMember(ctx context.Context, actorID, groupID, playerID int64) error {
	if err := validateIDs(actorID, groupID, playerID); err != nil {
		return err
	}
	g, err := s.loadGroup(ctx, groupID)
	if err != nil {
		return err
	}
	actor, err := s.requireManager(ctx, g.ID, actorID)
	if err != nil {
		return err
	}
	target, err := s.membership(ctx, groupID, playerID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrGroupNotFound
	}
	if target.IsOwner() || playerID == g.OwnerPlayerID {
		return ErrGroupForbidden
	}
	if actor.Role == entity.GroupRoleAdmin && target.Role != entity.GroupRoleMember {
		return ErrGroupForbidden
	}
	return s.groups.DeleteMembership(ctx, groupID, playerID)
}

func (s *Service) ListJoinRequests(ctx context.Context, actorID, groupID int64) ([]port.GroupMember, error) {
	if err := validateIDs(actorID, groupID); err != nil {
		return nil, err
	}
	if _, err := s.loadGroup(ctx, groupID); err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, groupID, actorID); err != nil {
		return nil, err
	}
	rows, err := s.groups.ListPendingMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupMember, len(rows))
	for i, row := range rows {
		out[i] = port.GroupMember{
			PlayerID:   row.Membership.PlayerID,
			PlayerName: row.PlayerName,
			Role:       row.Membership.Role,
			Status:     row.Membership.Status,
		}
	}
	return out, nil
}

func (s *Service) ApproveJoinRequest(ctx context.Context, actorID, groupID, playerID int64) (*entity.GroupMembership, error) {
	if err := validateIDs(actorID, groupID, playerID); err != nil {
		return nil, err
	}
	if _, err := s.loadGroup(ctx, groupID); err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, groupID, actorID); err != nil {
		return nil, err
	}
	m, err := s.membership(ctx, groupID, playerID)
	if err != nil {
		return nil, err
	}
	if m == nil || m.Status != entity.GroupMembershipPending {
		return nil, ErrGroupNotFound
	}
	m.Status = entity.GroupMembershipActive
	m.Role = entity.GroupRoleMember
	return s.groups.UpdateMembership(ctx, *m)
}

func (s *Service) DenyJoinRequest(ctx context.Context, actorID, groupID, playerID int64) error {
	if err := validateIDs(actorID, groupID, playerID); err != nil {
		return err
	}
	if _, err := s.loadGroup(ctx, groupID); err != nil {
		return err
	}
	if _, err := s.requireManager(ctx, groupID, actorID); err != nil {
		return err
	}
	m, err := s.membership(ctx, groupID, playerID)
	if err != nil {
		return err
	}
	if m == nil || m.Status != entity.GroupMembershipPending {
		return ErrGroupNotFound
	}
	return s.groups.DeleteMembership(ctx, groupID, playerID)
}

func invitationExpiry(now time.Time) *time.Time {
	t := now.Add(invitationTTLDays * 24 * time.Hour)
	return &t
}
