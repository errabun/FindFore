package events_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type fakeEventGroups struct {
	groups      map[int64]*entity.Group
	memberships map[string]*entity.GroupMembership
}

func newFakeEventGroups() *fakeEventGroups {
	return &fakeEventGroups{groups: map[int64]*entity.Group{}, memberships: map[string]*entity.GroupMembership{}}
}

func (f *fakeEventGroups) seed(g entity.Group, members ...entity.GroupMembership) {
	cp := g
	f.groups[g.ID] = &cp
	for _, m := range members {
		m := m
		f.memberships[fmt.Sprintf("%d/%d", m.GroupID, m.PlayerID)] = &m
	}
}

func (f *fakeEventGroups) GetByID(_ context.Context, id int64) (*entity.Group, error) {
	g, ok := f.groups[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *g
	return &cp, nil
}
func (f *fakeEventGroups) GetMembership(_ context.Context, groupID, playerID int64) (*entity.GroupMembership, error) {
	m, ok := f.memberships[fmt.Sprintf("%d/%d", groupID, playerID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *m
	return &cp, nil
}

func (f *fakeEventGroups) CreateWithOwner(context.Context, entity.Group) (*entity.Group, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) Update(context.Context, entity.Group) (*entity.Group, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) Delete(_ context.Context, id int64) error {
	delete(f.groups, id)
	for k, m := range f.memberships {
		if m.GroupID == id {
			delete(f.memberships, k)
		}
	}
	return nil
}
func (f *fakeEventGroups) TransferOwnership(context.Context, int64, int64, int64) error {
	return nil
}
func (f *fakeEventGroups) ListPublic(context.Context, string, int32, int32) ([]entity.Group, error) {
	return nil, nil
}
func (f *fakeEventGroups) ListByPlayer(_ context.Context, playerID int64, _, _ int32) ([]entity.Group, error) {
	var out []entity.Group
	for _, m := range f.memberships {
		if m.PlayerID == playerID && m.IsActive() {
			if g, ok := f.groups[m.GroupID]; ok {
				out = append(out, *g)
			}
		}
	}
	return out, nil
}
func (f *fakeEventGroups) ListPublicSummaries(context.Context, int64, string, int32, int32) ([]port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeEventGroups) ListByPlayerSummaries(context.Context, int64, int32, int32) ([]port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeEventGroups) CountActiveMembers(context.Context, int64) (int64, error) { return 0, nil }
func (f *fakeEventGroups) ListActiveMembers(context.Context, int64, int32, int32) ([]port.GroupMemberRow, error) {
	return nil, nil
}
func (f *fakeEventGroups) ListPendingMembers(context.Context, int64) ([]port.GroupMemberRow, error) {
	return nil, nil
}
func (f *fakeEventGroups) InsertMembership(context.Context, entity.GroupMembership) (*entity.GroupMembership, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) UpdateMembership(context.Context, entity.GroupMembership) (*entity.GroupMembership, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) DeleteMembership(_ context.Context, groupID, playerID int64) error {
	delete(f.memberships, fmt.Sprintf("%d/%d", groupID, playerID))
	return nil
}
func (f *fakeEventGroups) GetInvitationByID(context.Context, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) GetOutstandingInvitation(context.Context, int64, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) ListInvitationsByInvitee(context.Context, int64) ([]port.GroupInvitationRow, error) {
	return nil, nil
}
func (f *fakeEventGroups) ListOutstandingInvitations(context.Context, int64) ([]port.GroupInvitationRow, error) {
	return nil, nil
}
func (f *fakeEventGroups) InsertInvitation(context.Context, entity.GroupInvitation) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) MarkInvitationAccepted(context.Context, int64, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) MarkInvitationDeclined(context.Context, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeEventGroups) AcceptInvitation(context.Context, int64, int64) (*entity.GroupMembership, error) {
	return nil, sql.ErrNoRows
}

func seededGroupEvents() (*fakeEventRepo, *fakePlayerEventRepo, *fakeEventGroups, *events.Service) {
	eventRepo := newFakeEventRepo()
	playerEvents := newFakePlayerEventRepo()
	groups := newFakeEventGroups()
	groups.seed(
		entity.Group{ID: 1, OwnerPlayerID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic},
		entity.GroupMembership{GroupID: 1, PlayerID: 1, Role: entity.GroupRoleOwner, Status: entity.GroupMembershipActive},
		entity.GroupMembership{GroupID: 1, PlayerID: 2, Role: entity.GroupRoleMember, Status: entity.GroupMembershipActive},
		entity.GroupMembership{GroupID: 1, PlayerID: 3, Role: entity.GroupRoleMember, Status: entity.GroupMembershipPending},
	)
	svc := events.NewService(eventRepo, playerEvents, fakeCourseRepo{}, groups)
	eventRepo.groupMembers = groups
	eventRepo.activeMember = map[string]bool{}
	eventRepo.groupNames = map[int64]string{}
	for _, m := range groups.memberships {
		if m.IsActive() {
			eventRepo.activeMember[fmt.Sprintf("%d/%d", m.GroupID, m.PlayerID)] = true
		}
	}
	for _, g := range groups.groups {
		eventRepo.groupNames[g.ID] = g.Name
	}
	return eventRepo, playerEvents, groups, svc
}

func TestGroupEventCreateRequiresActiveMember(t *testing.T) {
	_, _, _, svc := seededGroupEvents()
	gid := int64(1)
	ctx := context.Background()

	created, err := svc.Create(ctx, entity.Event{
		CourseID: 1, Date: "2099-01-01", TeeTime: "08:00", OpenSpots: 4, NumberOfHoles: "18",
		HostID: 2, GroupID: &gid, Private: false,
	}, nil)
	require.NoError(t, err)
	require.True(t, created.Private, "group rounds must be private")
	require.NotNil(t, created.GroupID)
	require.Equal(t, int64(1), *created.GroupID)

	_, err = svc.Create(ctx, entity.Event{
		CourseID: 1, Date: "2099-01-01", TeeTime: "08:00", OpenSpots: 4, NumberOfHoles: "18",
		HostID: 4, GroupID: &gid,
	}, nil)
	require.ErrorIs(t, err, events.ErrEventNotFound)

	_, err = svc.Create(ctx, entity.Event{
		CourseID: 1, Date: "2099-01-01", TeeTime: "08:00", OpenSpots: 4, NumberOfHoles: "18",
		HostID: 3, GroupID: &gid,
	}, nil)
	require.ErrorIs(t, err, events.ErrEventNotFound)
}

func TestGroupEventListAndGetVisibility(t *testing.T) {
	eventRepo, _, _, svc := seededGroupEvents()
	gid := int64(1)
	starts := futureStarts()
	eventRepo.byID[20] = &entity.Event{ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4, PlannedStartsAt: starts}
	eventRepo.details[20] = &entity.EventWithDetails{
		ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4,
		PlannedStartsAt: starts, CourseTimezone: entity.DefaultCourseTimezone,
	}

	ctx := context.Background()

	listed, err := svc.ListForGroup(ctx, 2, 1)
	require.NoError(t, err)
	require.Len(t, listed, 1)

	_, err = svc.ListForGroup(ctx, 4, 1)
	require.ErrorIs(t, err, events.ErrEventNotFound)

	_, err = svc.Get(ctx, 20, 2)
	require.NoError(t, err)
	_, err = svc.Get(ctx, 20, 4)
	require.ErrorIs(t, err, events.ErrEventNotFound)

	public, err := svc.List(ctx, 4, nil, true)
	require.NoError(t, err)
	require.Empty(t, public)
}

func TestJoinableGroupRoundsNeedOneMore(t *testing.T) {
	eventRepo, playerEvents, _, svc := seededGroupEvents()
	gid := int64(1)
	starts := futureStarts()
	eventRepo.byID[20] = &entity.Event{ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4, PlannedStartsAt: starts}
	eventRepo.details[20] = &entity.EventWithDetails{
		ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4,
		PlannedStartsAt: starts, CourseTimezone: entity.DefaultCourseTimezone, CourseName: "Test Course",
	}
	playerEvents.byStatus[20] = map[entity.InviteStatus][]int64{
		entity.InviteStatusAccepted: {1},
	}

	fullID := int64(21)
	eventRepo.byID[fullID] = &entity.Event{ID: fullID, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 1, PlannedStartsAt: starts}
	eventRepo.details[fullID] = &entity.EventWithDetails{
		ID: fullID, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 1,
		PlannedStartsAt: starts, CourseTimezone: entity.DefaultCourseTimezone,
	}
	playerEvents.byStatus[fullID] = map[entity.InviteStatus][]int64{
		entity.InviteStatusAccepted: {1},
	}

	ctx := context.Background()

	listed, err := svc.ListJoinableFromGroups(ctx, 2)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, int64(20), listed[0].ID)
	require.Equal(t, "Crew", listed[0].GroupName)

	hostListed, err := svc.ListJoinableFromGroups(ctx, 1)
	require.NoError(t, err)
	require.Empty(t, hostListed)

	pendingListed, err := svc.ListJoinableFromGroups(ctx, 3)
	require.NoError(t, err)
	require.Empty(t, pendingListed)

	strangerListed, err := svc.ListJoinableFromGroups(ctx, 4)
	require.NoError(t, err)
	require.Empty(t, strangerListed)
}

func TestRemovedMemberRemainsOnExistingRound(t *testing.T) {
	eventRepo, playerEvents, groupRepo, eventSvc := seededGroupEvents()
	groupSvc, _ := newGroupServiceForEvents(groupRepo)
	gid := int64(1)
	starts := futureStarts()
	eventRepo.byID[20] = &entity.Event{ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4, PlannedStartsAt: starts}
	eventRepo.details[20] = &entity.EventWithDetails{
		ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4,
		PlannedStartsAt: starts, CourseTimezone: entity.DefaultCourseTimezone, CourseName: "Test Course",
	}
	playerEvents.byStatus[20] = map[entity.InviteStatus][]int64{
		entity.InviteStatusAccepted: {1, 2},
	}

	ctx := context.Background()
	require.NoError(t, groupSvc.RemoveMember(ctx, 1, 1, 2))

	_, err := eventSvc.ListForGroup(ctx, 2, 1)
	require.ErrorIs(t, err, events.ErrEventNotFound)

	got, err := eventSvc.Get(ctx, 20, 2)
	require.NoError(t, err)
	require.Contains(t, got.Accepted, int64(2))

	joinable, err := eventSvc.ListJoinableFromGroups(ctx, 2)
	require.NoError(t, err)
	require.Empty(t, joinable)
}

func TestGroupDeleteHidesRoundsFromGroupFeeds(t *testing.T) {
	eventRepo, playerEvents, groupRepo, eventSvc := seededGroupEvents()
	groupSvc, _ := newGroupServiceForEvents(groupRepo)
	gid := int64(1)
	starts := futureStarts()
	eventRepo.byID[20] = &entity.Event{ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4, PlannedStartsAt: starts}
	eventRepo.details[20] = &entity.EventWithDetails{
		ID: 20, HostID: 1, Private: true, GroupID: &gid, OpenSpots: 4,
		PlannedStartsAt: starts, CourseTimezone: entity.DefaultCourseTimezone,
	}
	playerEvents.byStatus[20] = map[entity.InviteStatus][]int64{
		entity.InviteStatusAccepted: {1},
	}

	ctx := context.Background()
	require.NoError(t, groupSvc.Delete(ctx, 1, 1))

	// Application fakes do not apply FK SET NULL; production SQL does (see fk_invariants_test).
	// After the group is gone, group feeds and discovery must not surface the round.
	_, err := eventSvc.ListForGroup(ctx, 1, 1)
	require.ErrorIs(t, err, events.ErrEventNotFound)
	joinable, err := eventSvc.ListJoinableFromGroups(ctx, 1)
	require.NoError(t, err)
	require.Empty(t, joinable)

	got, err := eventSvc.Get(ctx, 20, 1)
	require.NoError(t, err)
	require.Equal(t, int64(20), got.ID)
}

func newGroupServiceForEvents(repo *fakeEventGroups) (*groups.Service, *fakeEventGroups) {
	players := &eventPlayers{byID: map[int64]*entity.Player{
		1: {ID: 1, Name: "Eric"},
		2: {ID: 2, Name: "Sam"},
	}}
	return groups.NewService(repo, players), repo
}

type eventPlayers struct {
	byID map[int64]*entity.Player
}

func (f *eventPlayers) List(context.Context) ([]entity.Player, error) { return nil, nil }
func (f *eventPlayers) GetByID(_ context.Context, id int64) (*entity.Player, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}
func (f *eventPlayers) GetByEmail(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *eventPlayers) GetByUsername(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *eventPlayers) Create(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *eventPlayers) Update(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (f *eventPlayers) GetPasswordByID(context.Context, int64) (string, error) {
	return "", sql.ErrNoRows
}
func (f *eventPlayers) UpdatePassword(context.Context, int64, string) error   { return sql.ErrNoRows }
func (f *eventPlayers) GetTokenVersion(context.Context, int64) (int32, error) { return 0, nil }
func (f *eventPlayers) ListIDsExcept(context.Context, int64) ([]int64, error) { return nil, nil }
