package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type GroupRepo struct {
	q  *sqlcgen.Queries
	db *sql.DB
}

func NewGroupRepo(q *sqlcgen.Queries, db *sql.DB) *GroupRepo {
	return &GroupRepo{q: q, db: db}
}

func mapGroup(row sqlcgen.Group) entity.Group {
	return entity.Group{
		ID:            row.ID,
		OwnerPlayerID: row.OwnerPlayerID,
		Name:          row.Name,
		Description:   row.Description,
		Privacy:       row.Privacy,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func mapMembership(row sqlcgen.GroupMembership) entity.GroupMembership {
	return entity.GroupMembership{
		GroupID:   row.GroupID,
		PlayerID:  row.PlayerID,
		Role:      row.Role,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func mapInvitation(row sqlcgen.GroupInvitation) entity.GroupInvitation {
	return entity.GroupInvitation{
		ID:              row.ID,
		GroupID:         row.GroupID,
		InviterPlayerID: row.InviterPlayerID,
		InviteePlayerID: row.InviteePlayerID,
		CreatedAt:       row.CreatedAt,
		ExpiresAt:       timePtr(row.ExpiresAt),
		AcceptedAt:      timePtr(row.AcceptedAt),
		DeclinedAt:      timePtr(row.DeclinedAt),
	}
}

func (r *GroupRepo) GetByID(ctx context.Context, id int64) (*entity.Group, error) {
	row, err := r.q.GetGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}
	g := mapGroup(row)
	return &g, nil
}

func (r *GroupRepo) CreateWithOwner(ctx context.Context, g entity.Group) (*entity.Group, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin group tx: %w", err)
	}
	defer tx.Rollback()
	qtx := r.q.WithTx(tx)

	row, err := qtx.InsertGroup(ctx, sqlcgen.InsertGroupParams{
		OwnerPlayerID: g.OwnerPlayerID,
		Name:          g.Name,
		Description:   g.Description,
		Privacy:       g.Privacy,
	})
	if err != nil {
		return nil, err
	}
	_, err = qtx.InsertGroupMembership(ctx, sqlcgen.InsertGroupMembershipParams{
		GroupID:  row.ID,
		PlayerID: g.OwnerPlayerID,
		Role:     entity.GroupRoleOwner,
		Status:   entity.GroupMembershipActive,
	})
	if err != nil {
		return nil, fmt.Errorf("insert owner membership: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit group: %w", err)
	}
	out := mapGroup(row)
	return &out, nil
}

func (r *GroupRepo) Update(ctx context.Context, g entity.Group) (*entity.Group, error) {
	row, err := r.q.UpdateGroup(ctx, sqlcgen.UpdateGroupParams{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Privacy:     g.Privacy,
	})
	if err != nil {
		return nil, err
	}
	out := mapGroup(row)
	return &out, nil
}

func (r *GroupRepo) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteGroup(ctx, id)
}

func (r *GroupRepo) TransferOwnership(ctx context.Context, groupID, fromPlayerID, toPlayerID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transfer ownership tx: %w", err)
	}
	defer tx.Rollback()
	qtx := r.q.WithTx(tx)

	if err := qtx.UpdateGroupOwner(ctx, sqlcgen.UpdateGroupOwnerParams{
		ID: groupID, OwnerPlayerID: toPlayerID,
	}); err != nil {
		return err
	}
	from, err := qtx.GetGroupMembership(ctx, sqlcgen.GetGroupMembershipParams{
		GroupID: groupID, PlayerID: fromPlayerID,
	})
	if err != nil {
		return err
	}
	if _, err := qtx.UpdateGroupMembership(ctx, sqlcgen.UpdateGroupMembershipParams{
		GroupID: from.GroupID, PlayerID: from.PlayerID,
		Role: entity.GroupRoleMember, Status: from.Status,
	}); err != nil {
		return err
	}
	to, err := qtx.GetGroupMembership(ctx, sqlcgen.GetGroupMembershipParams{
		GroupID: groupID, PlayerID: toPlayerID,
	})
	if err != nil {
		return err
	}
	if _, err := qtx.UpdateGroupMembership(ctx, sqlcgen.UpdateGroupMembershipParams{
		GroupID: to.GroupID, PlayerID: to.PlayerID,
		Role: entity.GroupRoleOwner, Status: to.Status,
	}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transfer ownership: %w", err)
	}
	return nil
}

func (r *GroupRepo) ListPublic(ctx context.Context, search string, limit, offset int32) ([]entity.Group, error) {
	rows, err := r.q.ListPublicGroups(ctx, sqlcgen.ListPublicGroupsParams{
		Search: search,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entity.Group, len(rows))
	for i, row := range rows {
		out[i] = mapGroup(row)
	}
	return out, nil
}

func (r *GroupRepo) ListByPlayer(ctx context.Context, playerID int64, limit, offset int32) ([]entity.Group, error) {
	rows, err := r.q.ListGroupsByPlayer(ctx, sqlcgen.ListGroupsByPlayerParams{
		PlayerID: playerID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entity.Group, len(rows))
	for i, row := range rows {
		out[i] = mapGroup(row)
	}
	return out, nil
}

func mapSummaryGroup(id, ownerPlayerID int64, name, description, privacy string, createdAt, updatedAt time.Time) entity.Group {
	return entity.Group{
		ID:            id,
		OwnerPlayerID: ownerPlayerID,
		Name:          name,
		Description:   description,
		Privacy:       privacy,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func viewerMembership(groupID, playerID int64, role, status string, createdAt, updatedAt time.Time) *entity.GroupMembership {
	return &entity.GroupMembership{
		GroupID:   groupID,
		PlayerID:  playerID,
		Role:      role,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (r *GroupRepo) ListPublicSummaries(ctx context.Context, playerID int64, search string, limit, offset int32) ([]port.GroupDetails, error) {
	rows, err := r.q.ListPublicGroupSummaries(ctx, sqlcgen.ListPublicGroupSummariesParams{
		PlayerID: playerID,
		Search:   search,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupDetails, len(rows))
	for i, row := range rows {
		d := port.GroupDetails{
			Group:       mapSummaryGroup(row.ID, row.OwnerPlayerID, row.Name, row.Description, row.Privacy, row.CreatedAt, row.UpdatedAt),
			OwnerName:   row.OwnerName,
			MemberCount: row.MemberCount,
		}
		if row.ViewerRole.Valid && row.ViewerStatus.Valid {
			d.Viewer = viewerMembership(row.ID, playerID, row.ViewerRole.String, row.ViewerStatus.String, row.ViewerCreatedAt.Time, row.ViewerUpdatedAt.Time)
		}
		out[i] = d
	}
	return out, nil
}

func (r *GroupRepo) ListByPlayerSummaries(ctx context.Context, playerID int64, limit, offset int32) ([]port.GroupDetails, error) {
	rows, err := r.q.ListGroupSummariesByPlayer(ctx, sqlcgen.ListGroupSummariesByPlayerParams{
		PlayerID: playerID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupDetails, len(rows))
	for i, row := range rows {
		out[i] = port.GroupDetails{
			Group:       mapSummaryGroup(row.ID, row.OwnerPlayerID, row.Name, row.Description, row.Privacy, row.CreatedAt, row.UpdatedAt),
			OwnerName:   row.OwnerName,
			MemberCount: row.MemberCount,
			Viewer:      viewerMembership(row.ID, playerID, row.ViewerRole, row.ViewerStatus, row.ViewerCreatedAt, row.ViewerUpdatedAt),
		}
	}
	return out, nil
}

func (r *GroupRepo) CountActiveMembers(ctx context.Context, groupID int64) (int64, error) {
	return r.q.CountActiveGroupMembers(ctx, groupID)
}

func (r *GroupRepo) GetMembership(ctx context.Context, groupID, playerID int64) (*entity.GroupMembership, error) {
	row, err := r.q.GetGroupMembership(ctx, sqlcgen.GetGroupMembershipParams{
		GroupID: groupID, PlayerID: playerID,
	})
	if err != nil {
		return nil, err
	}
	m := mapMembership(row)
	return &m, nil
}

func (r *GroupRepo) ListActiveMembers(ctx context.Context, groupID int64, limit, offset int32) ([]port.GroupMemberRow, error) {
	rows, err := r.q.ListActiveGroupMembers(ctx, sqlcgen.ListActiveGroupMembersParams{
		GroupID: groupID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupMemberRow, len(rows))
	for i, row := range rows {
		out[i] = port.GroupMemberRow{
			Membership: entity.GroupMembership{
				GroupID: row.GroupID, PlayerID: row.PlayerID, Role: row.Role, Status: row.Status,
				CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			},
			PlayerName: row.PlayerName.String,
		}
	}
	return out, nil
}

func (r *GroupRepo) ListPendingMembers(ctx context.Context, groupID int64) ([]port.GroupMemberRow, error) {
	rows, err := r.q.ListPendingGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupMemberRow, len(rows))
	for i, row := range rows {
		out[i] = port.GroupMemberRow{
			Membership: entity.GroupMembership{
				GroupID: row.GroupID, PlayerID: row.PlayerID, Role: row.Role, Status: row.Status,
				CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			},
			PlayerName: row.PlayerName.String,
		}
	}
	return out, nil
}

func (r *GroupRepo) InsertMembership(ctx context.Context, m entity.GroupMembership) (*entity.GroupMembership, error) {
	row, err := r.q.InsertGroupMembership(ctx, sqlcgen.InsertGroupMembershipParams{
		GroupID: m.GroupID, PlayerID: m.PlayerID, Role: m.Role, Status: m.Status,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, entity.ErrGroupConflict
		}
		return nil, err
	}
	out := mapMembership(row)
	return &out, nil
}

func (r *GroupRepo) UpdateMembership(ctx context.Context, m entity.GroupMembership) (*entity.GroupMembership, error) {
	row, err := r.q.UpdateGroupMembership(ctx, sqlcgen.UpdateGroupMembershipParams{
		GroupID: m.GroupID, PlayerID: m.PlayerID, Role: m.Role, Status: m.Status,
	})
	if err != nil {
		return nil, err
	}
	out := mapMembership(row)
	return &out, nil
}

func (r *GroupRepo) DeleteMembership(ctx context.Context, groupID, playerID int64) error {
	return r.q.DeleteGroupMembership(ctx, sqlcgen.DeleteGroupMembershipParams{
		GroupID: groupID, PlayerID: playerID,
	})
}

func (r *GroupRepo) GetInvitationByID(ctx context.Context, id int64) (*entity.GroupInvitation, error) {
	row, err := r.q.GetGroupInvitationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	inv := mapInvitation(row)
	return &inv, nil
}

func (r *GroupRepo) GetOutstandingInvitation(ctx context.Context, groupID, inviteeID int64) (*entity.GroupInvitation, error) {
	row, err := r.q.GetOutstandingGroupInvitation(ctx, sqlcgen.GetOutstandingGroupInvitationParams{
		GroupID: groupID, InviteePlayerID: inviteeID,
	})
	if err != nil {
		return nil, err
	}
	inv := mapInvitation(row)
	return &inv, nil
}

func (r *GroupRepo) ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]port.GroupInvitationRow, error) {
	rows, err := r.q.ListGroupInvitationsByInvitee(ctx, inviteeID)
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupInvitationRow, 0, len(rows))
	for _, row := range rows {
		inv := entity.GroupInvitation{
			ID: row.ID, GroupID: row.GroupID, InviterPlayerID: row.InviterPlayerID,
			InviteePlayerID: row.InviteePlayerID, CreatedAt: row.CreatedAt,
			ExpiresAt: timePtr(row.ExpiresAt), AcceptedAt: timePtr(row.AcceptedAt),
			DeclinedAt: timePtr(row.DeclinedAt),
		}
		out = append(out, port.GroupInvitationRow{
			Invitation: inv, GroupName: row.GroupName, InviterName: row.InviterName.String,
		})
	}
	return out, nil
}

func (r *GroupRepo) ListOutstandingInvitations(ctx context.Context, groupID int64) ([]port.GroupInvitationRow, error) {
	rows, err := r.q.ListOutstandingGroupInvitations(ctx, groupID)
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupInvitationRow, 0, len(rows))
	for _, row := range rows {
		inv := entity.GroupInvitation{
			ID: row.ID, GroupID: row.GroupID, InviterPlayerID: row.InviterPlayerID,
			InviteePlayerID: row.InviteePlayerID, CreatedAt: row.CreatedAt,
			ExpiresAt: timePtr(row.ExpiresAt), AcceptedAt: timePtr(row.AcceptedAt),
			DeclinedAt: timePtr(row.DeclinedAt),
		}
		out = append(out, port.GroupInvitationRow{
			Invitation: inv, GroupName: row.GroupName,
			InviterName: row.InviterName.String, InviteeName: row.InviteeName.String,
		})
	}
	return out, nil
}

func (r *GroupRepo) InsertInvitation(ctx context.Context, inv entity.GroupInvitation) (*entity.GroupInvitation, error) {
	row, err := r.q.InsertGroupInvitation(ctx, sqlcgen.InsertGroupInvitationParams{
		GroupID: inv.GroupID, InviterPlayerID: inv.InviterPlayerID,
		InviteePlayerID: inv.InviteePlayerID, ExpiresAt: nullTimePtr(inv.ExpiresAt),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, entity.ErrGroupConflict
		}
		return nil, err
	}
	out := mapInvitation(row)
	return &out, nil
}

func (r *GroupRepo) MarkInvitationAccepted(ctx context.Context, id, inviteeID int64) (*entity.GroupInvitation, error) {
	row, err := r.q.MarkGroupInvitationAccepted(ctx, sqlcgen.MarkGroupInvitationAcceptedParams{
		ID:              id,
		InviteePlayerID: inviteeID,
	})
	if err != nil {
		return nil, err
	}
	out := mapInvitation(row)
	return &out, nil
}

func (r *GroupRepo) MarkInvitationDeclined(ctx context.Context, id int64) (*entity.GroupInvitation, error) {
	row, err := r.q.MarkGroupInvitationDeclined(ctx, id)
	if err != nil {
		return nil, err
	}
	out := mapInvitation(row)
	return &out, nil
}

func (r *GroupRepo) AcceptInvitation(ctx context.Context, invitationID, playerID int64) (*entity.GroupMembership, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin accept invite tx: %w", err)
	}
	defer tx.Rollback()
	qtx := r.q.WithTx(tx)

	inv, err := qtx.MarkGroupInvitationAccepted(ctx, sqlcgen.MarkGroupInvitationAcceptedParams{
		ID:              invitationID,
		InviteePlayerID: playerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existingInv, getErr := qtx.GetGroupInvitationByID(ctx, invitationID)
			if getErr != nil {
				return nil, err
			}
			if existingInv.InviteePlayerID != playerID {
				return nil, entity.ErrGroupForbidden
			}
			if existingInv.AcceptedAt.Valid {
				existing, memErr := qtx.GetGroupMembership(ctx, sqlcgen.GetGroupMembershipParams{
					GroupID: existingInv.GroupID, PlayerID: playerID,
				})
				if memErr != nil {
					return nil, memErr
				}
				out := mapMembership(existing)
				return &out, nil
			}
			return nil, entity.ErrGroupConflict
		}
		return nil, err
	}

	existing, err := qtx.GetGroupMembership(ctx, sqlcgen.GetGroupMembershipParams{
		GroupID: inv.GroupID, PlayerID: playerID,
	})
	var membership sqlcgen.GroupMembership
	if err == sql.ErrNoRows {
		membership, err = qtx.InsertGroupMembership(ctx, sqlcgen.InsertGroupMembershipParams{
			GroupID: inv.GroupID, PlayerID: playerID,
			Role: entity.GroupRoleMember, Status: entity.GroupMembershipActive,
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		membership, err = qtx.UpdateGroupMembership(ctx, sqlcgen.UpdateGroupMembershipParams{
			GroupID: existing.GroupID, PlayerID: existing.PlayerID,
			Role: existing.Role, Status: entity.GroupMembershipActive,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit accept invite: %w", err)
	}
	out := mapMembership(membership)
	return &out, nil
}

var _ port.GroupRepository = (*GroupRepo)(nil)
