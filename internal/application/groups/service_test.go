package groups_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type fakePlayers struct {
	byID map[int64]*entity.Player
}

func newFakePlayers(ids ...int64) *fakePlayers {
	f := &fakePlayers{byID: map[int64]*entity.Player{}}
	for _, id := range ids {
		f.byID[id] = &entity.Player{ID: id, Name: "Player"}
	}
	return f
}

func (f *fakePlayers) List(context.Context) ([]entity.Player, error) { return nil, nil }
func (f *fakePlayers) GetByID(_ context.Context, id int64) (*entity.Player, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}
func (f *fakePlayers) GetByEmail(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *fakePlayers) GetByUsername(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *fakePlayers) Create(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *fakePlayers) Update(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *fakePlayers) GetPasswordByID(context.Context, int64) (string, error) {
	return "", sql.ErrNoRows
}
func (f *fakePlayers) UpdatePassword(context.Context, int64, string) error   { return sql.ErrNoRows }
func (f *fakePlayers) GetTokenVersion(context.Context, int64) (int32, error) { return 0, nil }
func (f *fakePlayers) ListIDsExcept(context.Context, int64) ([]int64, error) { return nil, nil }

type fakeGroups struct {
	groups      map[int64]*entity.Group
	memberships map[string]*entity.GroupMembership
	invitations map[int64]*entity.GroupInvitation
	nextGroup   int64
	nextInvite  int64
	playerNames map[int64]string
}

func newFakeGroups() *fakeGroups {
	return &fakeGroups{
		groups: map[int64]*entity.Group{}, memberships: map[string]*entity.GroupMembership{},
		invitations: map[int64]*entity.GroupInvitation{}, nextGroup: 1, nextInvite: 1,
		playerNames: map[int64]string{},
	}
}

func memKey(groupID, playerID int64) string {
	return itoa(groupID) + "/" + itoa(playerID)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func (f *fakeGroups) GetByID(_ context.Context, id int64) (*entity.Group, error) {
	g, ok := f.groups[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *g
	return &cp, nil
}

func (f *fakeGroups) CreateWithOwner(_ context.Context, g entity.Group) (*entity.Group, error) {
	g.ID = f.nextGroup
	f.nextGroup++
	now := time.Now().UTC()
	g.CreatedAt, g.UpdatedAt = now, now
	cp := g
	f.groups[g.ID] = &cp
	f.memberships[memKey(g.ID, g.OwnerPlayerID)] = &entity.GroupMembership{
		GroupID: g.ID, PlayerID: g.OwnerPlayerID,
		Role: entity.GroupRoleOwner, Status: entity.GroupMembershipActive,
		CreatedAt: now, UpdatedAt: now,
	}
	return &g, nil
}

func (f *fakeGroups) Update(_ context.Context, g entity.Group) (*entity.Group, error) {
	g.UpdatedAt = time.Now().UTC()
	cp := g
	f.groups[g.ID] = &cp
	return &g, nil
}

func (f *fakeGroups) Delete(_ context.Context, id int64) error {
	delete(f.groups, id)
	for k, m := range f.memberships {
		if m.GroupID == id {
			delete(f.memberships, k)
		}
	}
	for k, inv := range f.invitations {
		if inv.GroupID == id {
			delete(f.invitations, k)
		}
	}
	return nil
}

func (f *fakeGroups) TransferOwnership(_ context.Context, groupID, fromPlayerID, toPlayerID int64) error {
	g, ok := f.groups[groupID]
	if !ok {
		return sql.ErrNoRows
	}
	g.OwnerPlayerID = toPlayerID
	if from, ok := f.memberships[memKey(groupID, fromPlayerID)]; ok {
		from.Role = entity.GroupRoleMember
	}
	if to, ok := f.memberships[memKey(groupID, toPlayerID)]; ok {
		to.Role = entity.GroupRoleOwner
	}
	return nil
}

func (f *fakeGroups) ListPublic(_ context.Context, search string, limit, offset int32) ([]entity.Group, error) {
	var out []entity.Group
	for _, g := range f.groups {
		if g.Privacy == entity.GroupPrivacyPublic {
			out = append(out, *g)
		}
	}
	return page(out, limit, offset), nil
}

func (f *fakeGroups) ListByPlayer(_ context.Context, playerID int64, limit, offset int32) ([]entity.Group, error) {
	var out []entity.Group
	for _, m := range f.memberships {
		if m.PlayerID == playerID && m.Status == entity.GroupMembershipActive {
			if g, ok := f.groups[m.GroupID]; ok {
				out = append(out, *g)
			}
		}
	}
	return page(out, limit, offset), nil
}

func (f *fakeGroups) summary(g entity.Group, actorID int64) port.GroupDetails {
	d := port.GroupDetails{Group: g, OwnerName: f.playerNames[g.OwnerPlayerID]}
	for _, m := range f.memberships {
		if m.GroupID == g.ID && m.Status == entity.GroupMembershipActive {
			d.MemberCount++
		}
	}
	if m, ok := f.memberships[memKey(g.ID, actorID)]; ok {
		cp := *m
		d.Viewer = &cp
	}
	return d
}

func (f *fakeGroups) ListPublicSummaries(_ context.Context, playerID int64, search string, limit, offset int32) ([]port.GroupDetails, error) {
	list, err := f.ListPublic(context.Background(), search, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupDetails, len(list))
	for i, g := range list {
		out[i] = f.summary(g, playerID)
	}
	return out, nil
}

func (f *fakeGroups) ListByPlayerSummaries(_ context.Context, playerID int64, limit, offset int32) ([]port.GroupDetails, error) {
	list, err := f.ListByPlayer(context.Background(), playerID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]port.GroupDetails, len(list))
	for i, g := range list {
		out[i] = f.summary(g, playerID)
	}
	return out, nil
}

func page(in []entity.Group, limit, offset int32) []entity.Group {
	if offset >= int32(len(in)) {
		return nil
	}
	in = in[offset:]
	if limit > 0 && int32(len(in)) > limit {
		in = in[:limit]
	}
	return in
}

func (f *fakeGroups) CountActiveMembers(_ context.Context, groupID int64) (int64, error) {
	var n int64
	for _, m := range f.memberships {
		if m.GroupID == groupID && m.Status == entity.GroupMembershipActive {
			n++
		}
	}
	return n, nil
}

func (f *fakeGroups) GetMembership(_ context.Context, groupID, playerID int64) (*entity.GroupMembership, error) {
	m, ok := f.memberships[memKey(groupID, playerID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *m
	return &cp, nil
}

func (f *fakeGroups) ListActiveMembers(_ context.Context, groupID int64, limit, offset int32) ([]port.GroupMemberRow, error) {
	var out []port.GroupMemberRow
	for _, m := range f.memberships {
		if m.GroupID == groupID && m.Status == entity.GroupMembershipActive {
			out = append(out, port.GroupMemberRow{Membership: *m, PlayerName: f.playerNames[m.PlayerID]})
		}
	}
	return out, nil
}

func (f *fakeGroups) ListPendingMembers(_ context.Context, groupID int64) ([]port.GroupMemberRow, error) {
	var out []port.GroupMemberRow
	for _, m := range f.memberships {
		if m.GroupID == groupID && m.Status == entity.GroupMembershipPending {
			out = append(out, port.GroupMemberRow{Membership: *m, PlayerName: f.playerNames[m.PlayerID]})
		}
	}
	return out, nil
}

func (f *fakeGroups) InsertMembership(_ context.Context, m entity.GroupMembership) (*entity.GroupMembership, error) {
	key := memKey(m.GroupID, m.PlayerID)
	if _, ok := f.memberships[key]; ok {
		return nil, entity.ErrGroupConflict
	}
	now := time.Now().UTC()
	m.CreatedAt, m.UpdatedAt = now, now
	cp := m
	f.memberships[key] = &cp
	return &m, nil
}

func (f *fakeGroups) UpdateMembership(_ context.Context, m entity.GroupMembership) (*entity.GroupMembership, error) {
	key := memKey(m.GroupID, m.PlayerID)
	m.UpdatedAt = time.Now().UTC()
	cp := m
	f.memberships[key] = &cp
	return &m, nil
}

func (f *fakeGroups) DeleteMembership(_ context.Context, groupID, playerID int64) error {
	delete(f.memberships, memKey(groupID, playerID))
	return nil
}

func (f *fakeGroups) GetInvitationByID(_ context.Context, id int64) (*entity.GroupInvitation, error) {
	inv, ok := f.invitations[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *inv
	return &cp, nil
}

func (f *fakeGroups) GetOutstandingInvitation(_ context.Context, groupID, inviteeID int64) (*entity.GroupInvitation, error) {
	for _, inv := range f.invitations {
		if inv.GroupID == groupID && inv.InviteePlayerID == inviteeID && inv.AcceptedAt == nil && inv.DeclinedAt == nil {
			cp := *inv
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (f *fakeGroups) ListInvitationsByInvitee(_ context.Context, inviteeID int64) ([]port.GroupInvitationRow, error) {
	var out []port.GroupInvitationRow
	for _, inv := range f.invitations {
		if inv.InviteePlayerID == inviteeID && inv.AcceptedAt == nil && inv.DeclinedAt == nil {
			name := ""
			if g, ok := f.groups[inv.GroupID]; ok {
				name = g.Name
			}
			out = append(out, port.GroupInvitationRow{Invitation: *inv, GroupName: name, InviterName: "Owner"})
		}
	}
	return out, nil
}

func (f *fakeGroups) ListOutstandingInvitations(_ context.Context, groupID int64) ([]port.GroupInvitationRow, error) {
	var out []port.GroupInvitationRow
	for _, inv := range f.invitations {
		if inv.GroupID == groupID && inv.AcceptedAt == nil && inv.DeclinedAt == nil {
			name := ""
			if g, ok := f.groups[inv.GroupID]; ok {
				name = g.Name
			}
			out = append(out, port.GroupInvitationRow{
				Invitation: *inv, GroupName: name, InviterName: "Owner",
				InviteeName: f.playerNames[inv.InviteePlayerID],
			})
		}
	}
	return out, nil
}

func (f *fakeGroups) InsertInvitation(_ context.Context, inv entity.GroupInvitation) (*entity.GroupInvitation, error) {
	if _, err := f.GetOutstandingInvitation(context.Background(), inv.GroupID, inv.InviteePlayerID); err == nil {
		return nil, entity.ErrGroupConflict
	}
	inv.ID = f.nextInvite
	f.nextInvite++
	inv.CreatedAt = time.Now().UTC()
	cp := inv
	f.invitations[inv.ID] = &cp
	return &inv, nil
}

func (f *fakeGroups) MarkInvitationAccepted(_ context.Context, id int64) (*entity.GroupInvitation, error) {
	inv, ok := f.invitations[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	now := time.Now().UTC()
	inv.AcceptedAt = &now
	cp := *inv
	return &cp, nil
}

func (f *fakeGroups) MarkInvitationDeclined(_ context.Context, id int64) (*entity.GroupInvitation, error) {
	inv, ok := f.invitations[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	now := time.Now().UTC()
	inv.DeclinedAt = &now
	cp := *inv
	return &cp, nil
}

func (f *fakeGroups) AcceptInvitation(ctx context.Context, invitationID, playerID int64) (*entity.GroupMembership, error) {
	inv, err := f.MarkInvitationAccepted(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if inv.InviteePlayerID != playerID {
		return nil, entity.ErrGroupForbidden
	}
	if existing, err := f.GetMembership(ctx, inv.GroupID, playerID); err == nil {
		existing.Status = entity.GroupMembershipActive
		return f.UpdateMembership(ctx, *existing)
	}
	return f.InsertMembership(ctx, entity.GroupMembership{
		GroupID: inv.GroupID, PlayerID: playerID,
		Role: entity.GroupRoleMember, Status: entity.GroupMembershipActive,
	})
}

func newSvc() (*groups.Service, *fakeGroups) {
	repo := newFakeGroups()
	players := newFakePlayers(1, 2, 3, 4)
	players.byID[1].Name = "Eric"
	players.byID[2].Name = "Sam"
	repo.playerNames[1] = "Eric"
	repo.playerNames[2] = "Sam"
	return groups.NewService(repo, players), repo
}

func TestCreateGroupOwnerMembership(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Saturday Morning Golf", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), d.Group.OwnerPlayerID)
	require.Equal(t, int64(1), d.MemberCount)
	require.NotNil(t, d.Viewer)
	require.Equal(t, entity.GroupRoleOwner, d.Viewer.Role)
}

func TestListMineAndDiscoverSummaries(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	created, err := svc.Create(ctx, port.CreateGroupInput{
		ActorID: 1, Name: "Saturday Morning Golf", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.Join(ctx, 2, created.Group.ID)
	require.NoError(t, err)

	mine, err := svc.ListMine(ctx, 1, 20, 0)
	require.NoError(t, err)
	require.Len(t, mine, 1)
	require.Equal(t, "Eric", mine[0].OwnerName)
	require.Equal(t, int64(2), mine[0].MemberCount)
	require.NotNil(t, mine[0].Viewer)
	require.Equal(t, entity.GroupRoleOwner, mine[0].Viewer.Role)

	discover, err := svc.ListDiscover(ctx, 2, "", 20, 0)
	require.NoError(t, err)
	require.Len(t, discover, 1)
	require.Equal(t, "Eric", discover[0].OwnerName)
	require.Equal(t, int64(2), discover[0].MemberCount)
	require.NotNil(t, discover[0].Viewer)
	require.Equal(t, entity.GroupRoleMember, discover[0].Viewer.Role)
	require.Equal(t, entity.GroupMembershipActive, discover[0].Viewer.Status)
}

func TestJoinPublicIdempotent(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Public Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)

	m1, err := svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	require.Equal(t, entity.GroupMembershipActive, m1.Status)
	require.Equal(t, entity.GroupRoleMember, m1.Role)

	m2, err := svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	require.Equal(t, m1.PlayerID, m2.PlayerID)
	require.Equal(t, entity.GroupMembershipActive, m2.Status)
}

func TestJoinPrivatePendingThenApprove(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Private Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)

	m, err := svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	require.Equal(t, entity.GroupMembershipPending, m.Status)

	approved, err := svc.ApproveJoinRequest(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)
	require.Equal(t, entity.GroupMembershipActive, approved.Status)
}

func TestDenyJoinRequestAllowsRetry(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Private Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	require.NoError(t, svc.DenyJoinRequest(context.Background(), 1, d.Group.ID, 2))

	m, err := svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	require.Equal(t, entity.GroupMembershipPending, m.Status)
}

func TestOwnerCannotLeave(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	err = svc.Leave(context.Background(), 1, d.Group.ID)
	require.ErrorIs(t, err, groups.ErrGroupOwnerCannotLeave)
}

func TestMemberCannotUpdateOrRemove(t *testing.T) {
	svc, repo := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 3, d.Group.ID)
	require.NoError(t, err)

	_, err = svc.Update(context.Background(), port.UpdateGroupInput{
		ActorID: 2, GroupID: d.Group.ID, Name: "Hacked", Privacy: entity.GroupPrivacyPublic,
	})
	require.ErrorIs(t, err, groups.ErrGroupForbidden)

	err = svc.RemoveMember(context.Background(), 2, d.Group.ID, 3)
	require.ErrorIs(t, err, groups.ErrGroupForbidden)
	_, err = repo.GetMembership(context.Background(), d.Group.ID, 3)
	require.NoError(t, err)
}

func TestInviteAcceptDecline(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)

	inv, err := svc.Invite(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)
	require.NotNil(t, inv)

	again, err := svc.Invite(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)
	require.Equal(t, inv.ID, again.ID)

	m, err := svc.AcceptInvitation(context.Background(), 2, inv.ID)
	require.NoError(t, err)
	require.Equal(t, entity.GroupMembershipActive, m.Status)

	_, err = svc.Invite(context.Background(), 1, d.Group.ID, 2)
	require.ErrorIs(t, err, groups.ErrGroupConflict)
}

func TestJoinAcceptsOutstandingInvite(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)
	_, err = svc.Invite(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)

	m, err := svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	require.Equal(t, entity.GroupMembershipActive, m.Status)
}

func TestInviteApprovesPendingRequest(t *testing.T) {
	svc, repo := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)

	inv, err := svc.Invite(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)
	require.Nil(t, inv)
	m, err := repo.GetMembership(context.Background(), d.Group.ID, 2)
	require.NoError(t, err)
	require.Equal(t, entity.GroupMembershipActive, m.Status)
}

func TestPrivateGroupHiddenFromStranger(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Secret", Privacy: entity.GroupPrivacyPrivate,
	})
	require.NoError(t, err)
	_, err = svc.Get(context.Background(), 2, d.Group.ID)
	require.ErrorIs(t, err, groups.ErrGroupNotFound)
	_, err = svc.ListMembers(context.Background(), 2, d.Group.ID, 20, 0)
	require.ErrorIs(t, err, groups.ErrGroupNotFound)
}

func TestDeclineInvitation(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	inv, err := svc.Invite(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)
	require.NoError(t, svc.DeclineInvitation(context.Background(), 2, inv.ID))
	require.NoError(t, svc.DeclineInvitation(context.Background(), 2, inv.ID))
}

func TestCancelInvitation(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	inv, err := svc.Invite(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)

	listed, err := svc.ListGroupInvitations(context.Background(), 1, d.Group.ID)
	require.NoError(t, err)
	require.Len(t, listed, 1)

	require.NoError(t, svc.CancelInvitation(context.Background(), 1, d.Group.ID, inv.ID))
	listed, err = svc.ListGroupInvitations(context.Background(), 1, d.Group.ID)
	require.NoError(t, err)
	require.Empty(t, listed)

	_, err = svc.AcceptInvitation(context.Background(), 2, inv.ID)
	require.ErrorIs(t, err, groups.ErrInvitationExpired)
}

func TestMemberCannotListInvitations(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	_, err = svc.ListGroupInvitations(context.Background(), 2, d.Group.ID)
	require.ErrorIs(t, err, groups.ErrGroupForbidden)
}

func TestTransferOwnershipThenLeave(t *testing.T) {
	svc, repo := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)

	updated, err := svc.TransferOwnership(context.Background(), 1, d.Group.ID, 2)
	require.NoError(t, err)
	require.Equal(t, int64(2), updated.Group.OwnerPlayerID)
	require.Equal(t, entity.GroupRoleMember, updated.Viewer.Role)

	require.NoError(t, svc.Leave(context.Background(), 1, d.Group.ID))
	_, err = repo.GetMembership(context.Background(), d.Group.ID, 1)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestTransferOwnershipRejectsNonMember(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.TransferOwnership(context.Background(), 1, d.Group.ID, 2)
	require.ErrorIs(t, err, groups.ErrInvalidGroup)
}

func TestNonOwnerCannotDeleteOrTransfer(t *testing.T) {
	svc, repo := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)

	err = svc.Delete(context.Background(), 2, d.Group.ID)
	require.ErrorIs(t, err, groups.ErrGroupForbidden)
	_, err = svc.TransferOwnership(context.Background(), 2, d.Group.ID, 1)
	require.ErrorIs(t, err, groups.ErrGroupForbidden)
	_, err = repo.GetByID(context.Background(), d.Group.ID)
	require.NoError(t, err)
}

func TestOwnerCanDelete(t *testing.T) {
	svc, repo := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), 1, d.Group.ID))
	_, err = repo.GetByID(context.Background(), d.Group.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestMemberCannotInvite(t *testing.T) {
	svc, _ := newSvc()
	d, err := svc.Create(context.Background(), port.CreateGroupInput{
		ActorID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic,
	})
	require.NoError(t, err)
	_, err = svc.Join(context.Background(), 2, d.Group.ID)
	require.NoError(t, err)
	_, err = svc.Invite(context.Background(), 2, d.Group.ID, 3)
	require.ErrorIs(t, err, groups.ErrGroupForbidden)
}
