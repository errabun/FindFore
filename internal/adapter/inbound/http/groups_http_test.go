package httphandler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	httphandler "github.com/ericrabun/findfore-go/internal/adapter/inbound/http"
	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type httpFakeGroups struct {
	groups      map[int64]*entity.Group
	memberships map[string]*entity.GroupMembership
	invitations map[int64]*entity.GroupInvitation
	nextGroup   int64
	nextInvite  int64
}

func newHTTPFakeGroups() *httpFakeGroups {
	return &httpFakeGroups{
		groups: map[int64]*entity.Group{}, memberships: map[string]*entity.GroupMembership{},
		invitations: map[int64]*entity.GroupInvitation{}, nextGroup: 1, nextInvite: 1,
	}
}

func gkey(groupID, playerID int64) string { return fmt.Sprintf("%d/%d", groupID, playerID) }

func (f *httpFakeGroups) GetByID(_ context.Context, id int64) (*entity.Group, error) {
	g, ok := f.groups[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *g
	return &cp, nil
}
func (f *httpFakeGroups) CreateWithOwner(_ context.Context, g entity.Group) (*entity.Group, error) {
	g.ID = f.nextGroup
	f.nextGroup++
	now := time.Now().UTC()
	g.CreatedAt, g.UpdatedAt = now, now
	cp := g
	f.groups[g.ID] = &cp
	f.memberships[gkey(g.ID, g.OwnerPlayerID)] = &entity.GroupMembership{
		GroupID: g.ID, PlayerID: g.OwnerPlayerID, Role: entity.GroupRoleOwner, Status: entity.GroupMembershipActive,
	}
	return &g, nil
}
func (f *httpFakeGroups) Update(_ context.Context, g entity.Group) (*entity.Group, error) {
	cp := g
	f.groups[g.ID] = &cp
	return &g, nil
}
func (f *httpFakeGroups) Delete(_ context.Context, id int64) error {
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
func (f *httpFakeGroups) TransferOwnership(_ context.Context, groupID, fromPlayerID, toPlayerID int64) error {
	g, ok := f.groups[groupID]
	if !ok {
		return sql.ErrNoRows
	}
	g.OwnerPlayerID = toPlayerID
	if from, ok := f.memberships[gkey(groupID, fromPlayerID)]; ok {
		from.Role = entity.GroupRoleMember
	}
	if to, ok := f.memberships[gkey(groupID, toPlayerID)]; ok {
		to.Role = entity.GroupRoleOwner
	}
	return nil
}
func (f *httpFakeGroups) ListPublic(_ context.Context, _ string, _, _ int32) ([]entity.Group, error) {
	var out []entity.Group
	for _, g := range f.groups {
		if g.IsPublic() {
			out = append(out, *g)
		}
	}
	return out, nil
}
func (f *httpFakeGroups) ListByPlayer(_ context.Context, playerID int64, _, _ int32) ([]entity.Group, error) {
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

func (f *httpFakeGroups) summary(g entity.Group, actorID int64) port.GroupDetails {
	d := port.GroupDetails{Group: g, OwnerName: "Player"}
	for _, m := range f.memberships {
		if m.GroupID == g.ID && m.IsActive() {
			d.MemberCount++
		}
	}
	if m, ok := f.memberships[gkey(g.ID, actorID)]; ok {
		cp := *m
		d.Viewer = &cp
	}
	return d
}

func (f *httpFakeGroups) ListPublicSummaries(_ context.Context, playerID int64, search string, limit, offset int32) ([]port.GroupDetails, error) {
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

func (f *httpFakeGroups) ListByPlayerSummaries(_ context.Context, playerID int64, limit, offset int32) ([]port.GroupDetails, error) {
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
func (f *httpFakeGroups) CountActiveMembers(_ context.Context, groupID int64) (int64, error) {
	var n int64
	for _, m := range f.memberships {
		if m.GroupID == groupID && m.IsActive() {
			n++
		}
	}
	return n, nil
}
func (f *httpFakeGroups) GetMembership(_ context.Context, groupID, playerID int64) (*entity.GroupMembership, error) {
	m, ok := f.memberships[gkey(groupID, playerID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *m
	return &cp, nil
}
func (f *httpFakeGroups) ListActiveMembers(_ context.Context, groupID int64, _, _ int32) ([]port.GroupMemberRow, error) {
	var out []port.GroupMemberRow
	for _, m := range f.memberships {
		if m.GroupID == groupID && m.IsActive() {
			out = append(out, port.GroupMemberRow{Membership: *m, PlayerName: "Player"})
		}
	}
	return out, nil
}
func (f *httpFakeGroups) ListPendingMembers(_ context.Context, groupID int64) ([]port.GroupMemberRow, error) {
	var out []port.GroupMemberRow
	for _, m := range f.memberships {
		if m.GroupID == groupID && m.Status == entity.GroupMembershipPending {
			out = append(out, port.GroupMemberRow{Membership: *m, PlayerName: "Player"})
		}
	}
	return out, nil
}
func (f *httpFakeGroups) InsertMembership(_ context.Context, m entity.GroupMembership) (*entity.GroupMembership, error) {
	key := gkey(m.GroupID, m.PlayerID)
	if _, ok := f.memberships[key]; ok {
		return nil, entity.ErrGroupConflict
	}
	cp := m
	f.memberships[key] = &cp
	return &m, nil
}
func (f *httpFakeGroups) UpdateMembership(_ context.Context, m entity.GroupMembership) (*entity.GroupMembership, error) {
	cp := m
	f.memberships[gkey(m.GroupID, m.PlayerID)] = &cp
	return &m, nil
}
func (f *httpFakeGroups) DeleteMembership(_ context.Context, groupID, playerID int64) error {
	delete(f.memberships, gkey(groupID, playerID))
	return nil
}
func (f *httpFakeGroups) GetInvitationByID(_ context.Context, id int64) (*entity.GroupInvitation, error) {
	inv, ok := f.invitations[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *inv
	return &cp, nil
}
func (f *httpFakeGroups) GetOutstandingInvitation(_ context.Context, groupID, inviteeID int64) (*entity.GroupInvitation, error) {
	for _, inv := range f.invitations {
		if inv.GroupID == groupID && inv.InviteePlayerID == inviteeID && inv.AcceptedAt == nil && inv.DeclinedAt == nil {
			cp := *inv
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}
func (f *httpFakeGroups) ListInvitationsByInvitee(_ context.Context, inviteeID int64) ([]port.GroupInvitationRow, error) {
	var out []port.GroupInvitationRow
	for _, inv := range f.invitations {
		if inv.InviteePlayerID == inviteeID && inv.AcceptedAt == nil && inv.DeclinedAt == nil {
			name := ""
			if g, ok := f.groups[inv.GroupID]; ok {
				name = g.Name
			}
			out = append(out, port.GroupInvitationRow{Invitation: *inv, GroupName: name, InviterName: "Eric"})
		}
	}
	return out, nil
}
func (f *httpFakeGroups) ListOutstandingInvitations(_ context.Context, groupID int64) ([]port.GroupInvitationRow, error) {
	var out []port.GroupInvitationRow
	for _, inv := range f.invitations {
		if inv.GroupID == groupID && inv.AcceptedAt == nil && inv.DeclinedAt == nil {
			name := ""
			if g, ok := f.groups[inv.GroupID]; ok {
				name = g.Name
			}
			out = append(out, port.GroupInvitationRow{
				Invitation: *inv, GroupName: name, InviterName: "Eric", InviteeName: "Player",
			})
		}
	}
	return out, nil
}
func (f *httpFakeGroups) InsertInvitation(_ context.Context, inv entity.GroupInvitation) (*entity.GroupInvitation, error) {
	inv.ID = f.nextInvite
	f.nextInvite++
	cp := inv
	f.invitations[inv.ID] = &cp
	return &inv, nil
}
func (f *httpFakeGroups) MarkInvitationAccepted(_ context.Context, id int64) (*entity.GroupInvitation, error) {
	now := time.Now().UTC()
	f.invitations[id].AcceptedAt = &now
	cp := *f.invitations[id]
	return &cp, nil
}
func (f *httpFakeGroups) MarkInvitationDeclined(_ context.Context, id int64) (*entity.GroupInvitation, error) {
	now := time.Now().UTC()
	f.invitations[id].DeclinedAt = &now
	cp := *f.invitations[id]
	return &cp, nil
}
func (f *httpFakeGroups) AcceptInvitation(ctx context.Context, invitationID, playerID int64) (*entity.GroupMembership, error) {
	inv, err := f.MarkInvitationAccepted(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	return f.InsertMembership(ctx, entity.GroupMembership{
		GroupID: inv.GroupID, PlayerID: playerID, Role: entity.GroupRoleMember, Status: entity.GroupMembershipActive,
	})
}

type groupsHTTPEnv struct {
	router http.Handler
	repo   *httpFakeGroups
}

func newGroupsHTTPEnv(t *testing.T) *groupsHTTPEnv {
	t.Helper()
	repo := newHTTPFakeGroups()
	svc := groups.NewService(repo, httpFakePlayers{})
	h := httphandler.New(stubPlayers{}, stubSessions{}, stubCourses{}, stubEvents{}, stubPlayerEvents{}, stubFriendships{}, stubPosts{}, nil, svc)
	return &groupsHTTPEnv{
		router: httphandler.NewRouter(h, testJWTSecret, stubTokenVersions{versions: map[int64]int32{1: 0, 2: 0, 3: 0}}),
		repo:   repo,
	}
}

func (e *groupsHTTPEnv) do(t *testing.T, method, path, body string, playerID int64) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	if playerID > 0 {
		tok, err := auth.GenerateToken(playerID, 0, testJWTSecret)
		require.NoError(t, err)
		r.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, r)
	return rec
}

func TestGroupsHTTPCreateAndOwnerMembership(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	rec := e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Saturday Morning Golf","privacy":"public"}`, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	var g struct {
		ID          int64 `json:"id"`
		MemberCount int64 `json:"member_count"`
		Viewer      *struct {
			Role   string `json:"role"`
			Status string `json:"status"`
		} `json:"viewer_membership"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &g))
	require.Equal(t, int64(1), g.MemberCount)
	require.Equal(t, "owner", g.Viewer.Role)
	require.Equal(t, "active", g.Viewer.Status)
}

func TestGroupsHTTPUnauthorizedUpdate(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	rec := e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	rec = e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)
	require.Equal(t, http.StatusOK, rec.Code)
	rec = e.do(t, http.MethodPatch, "/api/v1/groups/1", `{"name":"Hacked","privacy":"public"}`, 2)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGroupsHTTPPublicJoinIdempotent(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	rec := e.do(t, http.MethodPost, "/api/v1/groups/1/join", `{"role":"owner"}`, 2)
	require.Equal(t, http.StatusOK, rec.Code)
	var m struct {
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &m))
	require.Equal(t, "member", m.Role)
	require.Equal(t, "active", m.Status)
	rec = e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGroupsHTTPPrivateJoinPending(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"private"}`, 1)
	rec := e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)
	require.Equal(t, http.StatusOK, rec.Code)
	var m struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &m))
	require.Equal(t, "pending", m.Status)

	rec = e.do(t, http.MethodPost, "/api/v1/groups/1/join-requests/2/approve", "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGroupsHTTPInviteAccept(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"private"}`, 1)
	rec := e.do(t, http.MethodPost, "/api/v1/groups/1/invitations", `{"player_id":2}`, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	var inv struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &inv))
	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/group-invitations/%d/accept", inv.ID), "", 2)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGroupsHTTPMemberCannotRemove(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)
	e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 3)
	rec := e.do(t, http.MethodDelete, "/api/v1/groups/1/members/3", "", 2)
	require.Equal(t, http.StatusForbidden, rec.Code)
	_, err := e.repo.GetMembership(context.Background(), 1, 3)
	require.NoError(t, err)
}

func TestGroupsHTTPPrivateHidden(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Secret","privacy":"private"}`, 1)
	rec := e.do(t, http.MethodGet, "/api/v1/groups/1", "", 2)
	require.Equal(t, http.StatusNotFound, rec.Code)
	rec = e.do(t, http.MethodGet, "/api/v1/groups/1/members", "", 2)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGroupsHTTPOwnerCannotLeave(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	rec := e.do(t, http.MethodPost, "/api/v1/groups/1/leave", "", 1)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestGroupsHTTPMemberCannotDelete(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)
	rec := e.do(t, http.MethodDelete, "/api/v1/groups/1", "", 2)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGroupsHTTPTransferThenLeave(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)
	rec := e.do(t, http.MethodPost, "/api/v1/groups/1/transfer-ownership", `{"player_id":2}`, 1)
	require.Equal(t, http.StatusOK, rec.Code)
	rec = e.do(t, http.MethodPost, "/api/v1/groups/1/leave", "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGroupsHTTPCancelInvitation(t *testing.T) {
	e := newGroupsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"private"}`, 1)
	rec := e.do(t, http.MethodPost, "/api/v1/groups/1/invitations", `{"player_id":2}`, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	var inv struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &inv))

	rec = e.do(t, http.MethodGet, "/api/v1/groups/1/invitations", "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	rec = e.do(t, http.MethodGet, "/api/v1/groups/1/invitations", "", 2)
	require.Equal(t, http.StatusForbidden, rec.Code)

	rec = e.do(t, http.MethodDelete, fmt.Sprintf("/api/v1/groups/1/invitations/%d", inv.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	rec = e.do(t, http.MethodGet, "/api/v1/groups/1/invitations", "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Invitations []any `json:"invitations"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Empty(t, body.Invitations)
}
