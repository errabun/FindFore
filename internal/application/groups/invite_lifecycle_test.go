package groups_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func TestInvitationLifecycle(t *testing.T) {
	svc, repo := newSvc()
	ctx := context.Background()
	d, err := svc.Create(ctx, port.CreateGroupInput{
		ActorID: ownerID, Name: "Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)

	t.Run("invite self", func(t *testing.T) {
		_, err := svc.Invite(ctx, ownerID, d.Group.ID, ownerID)
		require.ErrorIs(t, err, groups.ErrInvalidGroup)
	})

	inv, err := svc.Invite(ctx, ownerID, d.Group.ID, memberID)
	require.NoError(t, err)

	t.Run("duplicate outstanding invite is idempotent", func(t *testing.T) {
		again, err := svc.Invite(ctx, ownerID, d.Group.ID, memberID)
		require.NoError(t, err)
		require.Equal(t, inv.ID, again.ID)
	})

	t.Run("accept twice is idempotent", func(t *testing.T) {
		m1, err := svc.AcceptInvitation(ctx, memberID, inv.ID)
		require.NoError(t, err)
		require.Equal(t, entity.GroupMembershipActive, m1.Status)
		m2, err := svc.AcceptInvitation(ctx, memberID, inv.ID)
		require.NoError(t, err)
		require.Equal(t, m1.PlayerID, m2.PlayerID)
		require.Equal(t, entity.GroupMembershipActive, m2.Status)
	})

	t.Run("invite someone already in the group", func(t *testing.T) {
		_, err := svc.Invite(ctx, ownerID, d.Group.ID, memberID)
		require.ErrorIs(t, err, groups.ErrGroupConflict)
	})

	t.Run("accepting someone else's invitation", func(t *testing.T) {
		other, err := svc.Invite(ctx, ownerID, d.Group.ID, strangerID)
		require.NoError(t, err)
		_, err = svc.AcceptInvitation(ctx, memberID, other.ID)
		require.ErrorIs(t, err, groups.ErrGroupForbidden)
	})

	t.Run("decline twice then accept declined", func(t *testing.T) {
		declined, err := svc.Invite(ctx, ownerID, d.Group.ID, adminID)
		require.NoError(t, err)
		require.NoError(t, svc.DeclineInvitation(ctx, adminID, declined.ID))
		require.NoError(t, svc.DeclineInvitation(ctx, adminID, declined.ID))
		_, err = svc.AcceptInvitation(ctx, adminID, declined.ID)
		require.ErrorIs(t, err, groups.ErrGroupConflict)
	})

	t.Run("accept expired invitation", func(t *testing.T) {
		expired, err := svc.Invite(ctx, ownerID, d.Group.ID, pendingID)
		require.NoError(t, err)
		past := time.Now().UTC().Add(-time.Hour)
		stored, err := repo.GetInvitationByID(ctx, expired.ID)
		require.NoError(t, err)
		stored.ExpiresAt = &past
		repo.invitations[expired.ID] = stored
		_, err = svc.AcceptInvitation(ctx, pendingID, expired.ID)
		require.ErrorIs(t, err, groups.ErrInvitationExpired)
	})

	t.Run("invite after being removed", func(t *testing.T) {
		require.NoError(t, svc.RemoveMember(ctx, ownerID, d.Group.ID, memberID))
		reinvite, err := svc.Invite(ctx, ownerID, d.Group.ID, memberID)
		require.NoError(t, err)
		m, err := svc.AcceptInvitation(ctx, memberID, reinvite.ID)
		require.NoError(t, err)
		require.Equal(t, entity.GroupMembershipActive, m.Status)
	})
}
