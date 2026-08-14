package groups_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

const (
	ownerID    int64 = 1
	memberID   int64 = 2
	adminID    int64 = 3
	pendingID  int64 = 4
	strangerID int64 = 5
	admin2ID   int64 = 6
)

type matrixEnv struct {
	svc       *groups.Service
	repo      *fakeGroups
	publicID  int64
	privateID int64
}

func setupMatrix(t *testing.T) matrixEnv {
	t.Helper()
	svc, repo := newSvc()
	ctx := context.Background()

	pub, err := svc.Create(ctx, port.CreateGroupInput{
		ActorID: ownerID, Name: "Public Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	priv, err := svc.Create(ctx, port.CreateGroupInput{
		ActorID: ownerID, Name: "Private Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)

	_, err = svc.Join(ctx, memberID, pub.Group.ID)
	require.NoError(t, err)
	_, err = svc.Join(ctx, adminID, pub.Group.ID)
	require.NoError(t, err)
	_, err = svc.Join(ctx, admin2ID, pub.Group.ID)
	require.NoError(t, err)
	setRole(t, repo, pub.Group.ID, adminID, entity.GroupRoleAdmin)
	setRole(t, repo, pub.Group.ID, admin2ID, entity.GroupRoleAdmin)

	_, err = svc.Join(ctx, memberID, priv.Group.ID)
	require.NoError(t, err)
	_, err = svc.ApproveJoinRequest(ctx, ownerID, priv.Group.ID, memberID)
	require.NoError(t, err)
	_, err = svc.Join(ctx, adminID, priv.Group.ID)
	require.NoError(t, err)
	_, err = svc.ApproveJoinRequest(ctx, ownerID, priv.Group.ID, adminID)
	require.NoError(t, err)
	setRole(t, repo, priv.Group.ID, adminID, entity.GroupRoleAdmin)
	_, err = svc.Join(ctx, pendingID, priv.Group.ID)
	require.NoError(t, err)

	return matrixEnv{svc: svc, repo: repo, publicID: pub.Group.ID, privateID: priv.Group.ID}
}

func setRole(t *testing.T, repo *fakeGroups, groupID, playerID int64, role string) {
	t.Helper()
	m, err := repo.GetMembership(context.Background(), groupID, playerID)
	require.NoError(t, err)
	m.Role = role
	_, err = repo.UpdateMembership(context.Background(), *m)
	require.NoError(t, err)
}

func TestAuthorizationMatrix(t *testing.T) {
	ctx := context.Background()
	e := setupMatrix(t)

	t.Run("view public group", func(t *testing.T) {
		for _, actor := range []int64{strangerID, pendingID, memberID, adminID, ownerID} {
			_, err := e.svc.Get(ctx, actor, e.publicID)
			require.NoError(t, err, "actor %d", actor)
		}
	})

	t.Run("view private group", func(t *testing.T) {
		_, err := e.svc.Get(ctx, strangerID, e.privateID)
		require.ErrorIs(t, err, groups.ErrGroupNotFound)

		d, err := e.svc.Get(ctx, pendingID, e.privateID)
		require.NoError(t, err)
		require.Equal(t, entity.GroupMembershipPending, d.Viewer.Status)

		for _, actor := range []int64{memberID, adminID, ownerID} {
			_, err := e.svc.Get(ctx, actor, e.privateID)
			require.NoError(t, err, "actor %d", actor)
		}

		inv, err := e.svc.Invite(ctx, ownerID, e.privateID, strangerID)
		require.NoError(t, err)
		require.NotNil(t, inv)
		d, err = e.svc.Get(ctx, strangerID, e.privateID)
		require.NoError(t, err)
		require.Nil(t, d.Viewer)
		_, err = e.svc.ListMembers(ctx, strangerID, e.privateID, 20, 0)
		require.ErrorIs(t, err, groups.ErrGroupNotFound)
	})

	t.Run("view private members", func(t *testing.T) {
		for _, actor := range []int64{strangerID, pendingID} {
			_, err := e.svc.ListMembers(ctx, actor, e.privateID, 20, 0)
			require.ErrorIs(t, err, groups.ErrGroupNotFound, "actor %d", actor)
		}
		for _, actor := range []int64{memberID, adminID, ownerID} {
			_, err := e.svc.ListMembers(ctx, actor, e.privateID, 20, 0)
			require.NoError(t, err, "actor %d", actor)
		}
	})

	t.Run("invite", func(t *testing.T) {
		_, err := e.svc.Invite(ctx, strangerID, e.privateID, memberID)
		require.ErrorIs(t, err, groups.ErrGroupForbidden)
		_, err = e.svc.Invite(ctx, pendingID, e.privateID, memberID)
		require.ErrorIs(t, err, groups.ErrGroupForbidden)
		_, err = e.svc.Invite(ctx, memberID, e.privateID, strangerID)
		require.ErrorIs(t, err, groups.ErrGroupForbidden)
		inv, err := e.svc.Invite(ctx, adminID, e.publicID, strangerID)
		require.NoError(t, err)
		require.NotNil(t, inv)
		inv2, err := e.svc.Invite(ctx, ownerID, e.publicID, pendingID)
		require.NoError(t, err)
		require.NotNil(t, inv2)
	})

	t.Run("remove member", func(t *testing.T) {
		for _, actor := range []int64{strangerID, pendingID, memberID} {
			err := e.svc.RemoveMember(ctx, actor, e.publicID, memberID)
			require.ErrorIs(t, err, groups.ErrGroupForbidden, "actor %d", actor)
		}
		err := e.svc.RemoveMember(ctx, adminID, e.publicID, ownerID)
		require.ErrorIs(t, err, groups.ErrGroupForbidden)
		err = e.svc.RemoveMember(ctx, adminID, e.publicID, admin2ID)
		require.ErrorIs(t, err, groups.ErrGroupForbidden)

		extra, err := e.svc.Join(ctx, pendingID, e.publicID)
		require.NoError(t, err)
		require.Equal(t, entity.GroupMembershipActive, extra.Status)
		require.NoError(t, e.svc.RemoveMember(ctx, adminID, e.publicID, pendingID))
	})

	t.Run("change settings", func(t *testing.T) {
		for _, actor := range []int64{strangerID, pendingID, memberID} {
			_, err := e.svc.Update(ctx, port.UpdateGroupInput{
				ActorID: actor, GroupID: e.publicID, Name: "Hacked", Privacy: entity.GroupPrivacyPublic,
			})
			require.ErrorIs(t, err, groups.ErrGroupForbidden, "actor %d", actor)
		}
		_, err := e.svc.Update(ctx, port.UpdateGroupInput{
			ActorID: adminID, GroupID: e.publicID, Name: "Public Crew", Privacy: entity.GroupPrivacyPublic,
		})
		require.NoError(t, err)
		_, err = e.svc.Update(ctx, port.UpdateGroupInput{
			ActorID: ownerID, GroupID: e.publicID, Name: "Public Crew", Privacy: entity.GroupPrivacyPublic,
		})
		require.NoError(t, err)
	})

	t.Run("transfer ownership", func(t *testing.T) {
		for _, actor := range []int64{strangerID, pendingID, memberID, adminID} {
			_, err := e.svc.TransferOwnership(ctx, actor, e.publicID, ownerID)
			require.ErrorIs(t, err, groups.ErrGroupForbidden, "actor %d", actor)
		}
	})

	t.Run("delete group", func(t *testing.T) {
		for _, actor := range []int64{strangerID, pendingID, memberID, adminID} {
			err := e.svc.Delete(ctx, actor, e.privateID)
			require.ErrorIs(t, err, groups.ErrGroupForbidden, "actor %d", actor)
		}
		_, err := e.repo.GetByID(ctx, e.privateID)
		require.NoError(t, err)
	})
}

func TestDeleteGroupCascadesMembershipsAndInvitations(t *testing.T) {
	svc, repo := newSvc()
	ctx := context.Background()
	d, err := svc.Create(ctx, port.CreateGroupInput{
		ActorID: ownerID, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.Join(ctx, memberID, d.Group.ID)
	require.NoError(t, err)
	inv, err := svc.Invite(ctx, ownerID, d.Group.ID, strangerID)
	require.NoError(t, err)
	require.NotNil(t, inv)

	require.NoError(t, svc.Delete(ctx, ownerID, d.Group.ID))
	_, err = repo.GetByID(ctx, d.Group.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = repo.GetMembership(ctx, d.Group.ID, ownerID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = repo.GetMembership(ctx, d.Group.ID, memberID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = repo.GetInvitationByID(ctx, inv.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = svc.Get(ctx, ownerID, d.Group.ID)
	require.ErrorIs(t, err, groups.ErrGroupNotFound)
	_, err = svc.AcceptInvitation(ctx, strangerID, inv.ID)
	require.ErrorIs(t, err, groups.ErrInvitationNotFound)
}
