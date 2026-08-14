package groups

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func (s *Service) Delete(ctx context.Context, actorID, groupID int64) error {
	if err := validateIDs(actorID, groupID); err != nil {
		return err
	}
	g, err := s.loadGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if _, err := s.requireOwner(ctx, g.ID, actorID); err != nil {
		return err
	}
	return s.groups.Delete(ctx, groupID)
}

func (s *Service) TransferOwnership(ctx context.Context, actorID, groupID, newOwnerID int64) (*port.GroupDetails, error) {
	if err := validateIDs(actorID, groupID, newOwnerID); err != nil {
		return nil, err
	}
	if actorID == newOwnerID {
		return nil, ErrInvalidGroup
	}
	g, err := s.loadGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireOwner(ctx, g.ID, actorID); err != nil {
		return nil, err
	}
	target, err := s.membership(ctx, groupID, newOwnerID)
	if err != nil {
		return nil, err
	}
	if target == nil || !target.IsActive() {
		return nil, ErrInvalidGroup
	}
	if err := s.groups.TransferOwnership(ctx, groupID, actorID, newOwnerID); err != nil {
		return nil, err
	}
	updated, err := s.loadGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	viewer, err := s.membership(ctx, groupID, actorID)
	if err != nil {
		return nil, err
	}
	return s.details(ctx, updated, viewer)
}

func (s *Service) ListGroupInvitations(ctx context.Context, actorID, groupID int64) ([]port.GroupInvitationView, error) {
	if err := validateIDs(actorID, groupID); err != nil {
		return nil, err
	}
	if _, err := s.loadGroup(ctx, groupID); err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, groupID, actorID); err != nil {
		return nil, err
	}
	rows, err := s.groups.ListOutstandingInvitations(ctx, groupID)
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

func (s *Service) CancelInvitation(ctx context.Context, actorID, groupID, invitationID int64) error {
	if err := validateIDs(actorID, groupID, invitationID); err != nil {
		return err
	}
	if _, err := s.loadGroup(ctx, groupID); err != nil {
		return err
	}
	if _, err := s.requireManager(ctx, groupID, actorID); err != nil {
		return err
	}
	inv, err := s.groups.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvitationNotFound
		}
		return err
	}
	if inv.GroupID != groupID {
		return ErrInvitationNotFound
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

func invitationView(row port.GroupInvitationRow) port.GroupInvitationView {
	return port.GroupInvitationView{
		Invitation:  row.Invitation,
		GroupName:   row.GroupName,
		InviterName: row.InviterName,
		InviteeName: row.InviteeName,
	}
}
