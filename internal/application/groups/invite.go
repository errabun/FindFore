package groups

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func (s *Service) Invite(ctx context.Context, actorID, groupID, inviteeID int64) (*entity.GroupInvitation, error) {
	if err := validateIDs(actorID, groupID, inviteeID); err != nil {
		return nil, err
	}
	if actorID == inviteeID {
		return nil, ErrInvalidGroup
	}
	if _, err := s.players.GetByID(ctx, inviteeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidGroup
		}
		return nil, err
	}
	if _, err := s.loadGroup(ctx, groupID); err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, groupID, actorID); err != nil {
		return nil, err
	}

	existing, err := s.membership(ctx, groupID, inviteeID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.IsActive() {
		return nil, ErrGroupConflict
	}
	if existing != nil && existing.Status == entity.GroupMembershipPending {
		existing.Status = entity.GroupMembershipActive
		existing.Role = entity.GroupRoleMember
		if _, err := s.groups.UpdateMembership(ctx, *existing); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if outstanding, err := s.groups.GetOutstandingInvitation(ctx, groupID, inviteeID); err == nil {
		if outstanding.IsOutstanding(time.Now().UTC()) {
			return outstanding, nil
		}
		if _, err := s.groups.MarkInvitationDeclined(ctx, outstanding.ID); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	inv, err := s.groups.InsertInvitation(ctx, entity.GroupInvitation{
		GroupID:         groupID,
		InviterPlayerID: actorID,
		InviteePlayerID: inviteeID,
		ExpiresAt:       invitationExpiry(time.Now().UTC()),
	})
	if err != nil {
		if errors.Is(err, entity.ErrGroupConflict) {
			existingInv, getErr := s.groups.GetOutstandingInvitation(ctx, groupID, inviteeID)
			if getErr == nil {
				return existingInv, nil
			}
			return nil, ErrGroupConflict
		}
		return nil, err
	}
	return inv, nil
}

func (s *Service) ListMyInvitations(ctx context.Context, actorID int64) ([]port.GroupInvitationView, error) {
	if actorID <= 0 {
		return nil, ErrInvalidGroup
	}
	rows, err := s.groups.ListInvitationsByInvitee(ctx, actorID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	out := make([]port.GroupInvitationView, 0, len(rows))
	for _, row := range rows {
		if !row.Invitation.IsOutstanding(now) {
			continue
		}
		out = append(out, invitationView(row))
	}
	return out, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, actorID, invitationID int64) (*entity.GroupMembership, error) {
	if err := validateIDs(actorID, invitationID); err != nil {
		return nil, err
	}
	inv, err := s.groups.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	if inv.InviteePlayerID != actorID {
		return nil, ErrGroupForbidden
	}
	if !inv.IsOutstanding(time.Now().UTC()) {
		if inv.AcceptedAt != nil {
			m, mErr := s.membership(ctx, inv.GroupID, actorID)
			if mErr != nil {
				return nil, mErr
			}
			if m != nil {
				return m, nil
			}
		}
		return nil, ErrInvitationExpired
	}
	return s.groups.AcceptInvitation(ctx, invitationID, actorID)
}

func (s *Service) DeclineInvitation(ctx context.Context, actorID, invitationID int64) error {
	if err := validateIDs(actorID, invitationID); err != nil {
		return err
	}
	inv, err := s.groups.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvitationNotFound
		}
		return err
	}
	if inv.InviteePlayerID != actorID {
		return ErrGroupForbidden
	}
	if inv.DeclinedAt != nil {
		return nil
	}
	if inv.AcceptedAt != nil {
		return ErrGroupConflict
	}
	_, err = s.groups.MarkInvitationDeclined(ctx, invitationID)
	return err
}
