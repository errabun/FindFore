package feed_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/feed"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type fakePosts struct {
	byID   map[int64]*entity.PostWithDetails
	nextID int64
}

func newFakePosts() *fakePosts {
	return &fakePosts{byID: map[int64]*entity.PostWithDetails{}, nextID: 1}
}

func (f *fakePosts) GetByID(_ context.Context, id int64) (*entity.PostWithDetails, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (f *fakePosts) List(_ context.Context, _, _ int32) ([]entity.PostWithDetails, error) {
	var out []entity.PostWithDetails
	for _, p := range f.byID {
		if p.GroupID == nil {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (f *fakePosts) ListByGroupID(_ context.Context, groupID int64, _, _ int32) ([]entity.PostWithDetails, error) {
	var out []entity.PostWithDetails
	for _, p := range f.byID {
		if p.GroupID != nil && *p.GroupID == groupID {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (f *fakePosts) Create(_ context.Context, playerID int64, body string, groupID *int64) (int64, error) {
	id := f.nextID
	f.nextID++
	p := &entity.PostWithDetails{
		ID: id, PlayerID: playerID, PlayerName: "Player", Body: body,
		GroupID: groupID, CreatedAt: time.Now().UTC(),
	}
	f.byID[id] = p
	return id, nil
}

func (f *fakePosts) Delete(_ context.Context, id, playerID int64) error {
	p, ok := f.byID[id]
	if !ok || p.PlayerID != playerID {
		return sql.ErrNoRows
	}
	delete(f.byID, id)
	return nil
}

func (f *fakePosts) DeleteByID(_ context.Context, id int64) error {
	if _, ok := f.byID[id]; !ok {
		return sql.ErrNoRows
	}
	delete(f.byID, id)
	return nil
}

type fakeReactions struct {
	items []entity.Reaction
}

func (f *fakeReactions) ListByPostID(_ context.Context, postID int64) ([]entity.Reaction, error) {
	var out []entity.Reaction
	for _, r := range f.items {
		if r.PostID == postID {
			out = append(out, r)
		}
	}
	return out, nil
}
func (f *fakeReactions) Find(_ context.Context, postID, playerID int64, emoji string) (*entity.Reaction, error) {
	for i := range f.items {
		r := f.items[i]
		if r.PostID == postID && r.PlayerID == playerID && r.Emoji == emoji {
			return &r, nil
		}
	}
	return nil, sql.ErrNoRows
}
func (f *fakeReactions) Create(_ context.Context, postID, playerID int64, emoji string) (*entity.Reaction, error) {
	r := entity.Reaction{ID: int64(len(f.items) + 1), PostID: postID, PlayerID: playerID, Emoji: emoji}
	f.items = append(f.items, r)
	return &r, nil
}
func (f *fakeReactions) Delete(_ context.Context, postID, playerID int64, emoji string) error {
	kept := f.items[:0]
	for _, r := range f.items {
		if !(r.PostID == postID && r.PlayerID == playerID && r.Emoji == emoji) {
			kept = append(kept, r)
		}
	}
	f.items = kept
	return nil
}

type fakeReplies struct{}

func (fakeReplies) ListByPostID(context.Context, int64) ([]entity.Reply, error) { return nil, nil }
func (fakeReplies) Create(context.Context, int64, int64, string) (*entity.Reply, error) {
	return &entity.Reply{ID: 1}, nil
}
func (fakeReplies) Delete(context.Context, int64, int64) error { return nil }

type fakeFeedGroups struct {
	groups      map[int64]*entity.Group
	memberships map[string]*entity.GroupMembership
}

func newFakeFeedGroups() *fakeFeedGroups {
	return &fakeFeedGroups{groups: map[int64]*entity.Group{}, memberships: map[string]*entity.GroupMembership{}}
}

func (f *fakeFeedGroups) seed(g entity.Group, members ...entity.GroupMembership) {
	cp := g
	f.groups[g.ID] = &cp
	for _, m := range members {
		m := m
		f.memberships[gkey(m.GroupID, m.PlayerID)] = &m
	}
}

func gkey(groupID, playerID int64) string {
	return itoa(groupID) + "/" + itoa(playerID)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func (f *fakeFeedGroups) GetByID(_ context.Context, id int64) (*entity.Group, error) {
	g, ok := f.groups[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *g
	return &cp, nil
}
func (f *fakeFeedGroups) GetMembership(_ context.Context, groupID, playerID int64) (*entity.GroupMembership, error) {
	m, ok := f.memberships[gkey(groupID, playerID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *m
	return &cp, nil
}

func (f *fakeFeedGroups) CreateWithOwner(context.Context, entity.Group) (*entity.Group, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) Update(context.Context, entity.Group) (*entity.Group, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) Delete(context.Context, int64) error { return nil }
func (f *fakeFeedGroups) TransferOwnership(context.Context, int64, int64, int64) error {
	return nil
}
func (f *fakeFeedGroups) ListPublic(context.Context, string, int32, int32) ([]entity.Group, error) {
	return nil, nil
}
func (f *fakeFeedGroups) ListByPlayer(context.Context, int64, int32, int32) ([]entity.Group, error) {
	return nil, nil
}
func (f *fakeFeedGroups) ListPublicSummaries(context.Context, int64, string, int32, int32) ([]port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeFeedGroups) ListByPlayerSummaries(context.Context, int64, int32, int32) ([]port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeFeedGroups) CountActiveMembers(context.Context, int64) (int64, error) { return 0, nil }
func (f *fakeFeedGroups) ListActiveMembers(context.Context, int64, int32, int32) ([]port.GroupMemberRow, error) {
	return nil, nil
}
func (f *fakeFeedGroups) ListPendingMembers(context.Context, int64) ([]port.GroupMemberRow, error) {
	return nil, nil
}
func (f *fakeFeedGroups) InsertMembership(context.Context, entity.GroupMembership) (*entity.GroupMembership, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) UpdateMembership(context.Context, entity.GroupMembership) (*entity.GroupMembership, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) DeleteMembership(context.Context, int64, int64) error { return nil }
func (f *fakeFeedGroups) GetInvitationByID(context.Context, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) GetOutstandingInvitation(context.Context, int64, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) ListInvitationsByInvitee(context.Context, int64) ([]port.GroupInvitationRow, error) {
	return nil, nil
}
func (f *fakeFeedGroups) ListOutstandingInvitations(context.Context, int64) ([]port.GroupInvitationRow, error) {
	return nil, nil
}
func (f *fakeFeedGroups) InsertInvitation(context.Context, entity.GroupInvitation) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) MarkInvitationAccepted(context.Context, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) MarkInvitationDeclined(context.Context, int64) (*entity.GroupInvitation, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeFeedGroups) AcceptInvitation(context.Context, int64, int64) (*entity.GroupMembership, error) {
	return nil, sql.ErrNoRows
}

func newFeedSvc() (*feed.Service, *fakePosts, *fakeFeedGroups) {
	posts := newFakePosts()
	groups := newFakeFeedGroups()
	groups.seed(
		entity.Group{ID: 1, OwnerPlayerID: 1, Name: "Crew", Privacy: entity.GroupPrivacyPublic},
		entity.GroupMembership{GroupID: 1, PlayerID: 1, Role: entity.GroupRoleOwner, Status: entity.GroupMembershipActive},
		entity.GroupMembership{GroupID: 1, PlayerID: 2, Role: entity.GroupRoleMember, Status: entity.GroupMembershipActive},
		entity.GroupMembership{GroupID: 1, PlayerID: 3, Role: entity.GroupRoleMember, Status: entity.GroupMembershipPending},
	)
	svc := feed.NewService(posts, &fakeReactions{}, fakeReplies{}, groups)
	return svc, posts, groups
}

func TestCommunityListExcludesGroupPosts(t *testing.T) {
	svc, _, _ := newFeedSvc()
	_, err := svc.Create(context.Background(), 1, "community hello")
	require.NoError(t, err)
	_, err = svc.CreateForGroup(context.Background(), 1, 1, "group only")
	require.NoError(t, err)

	list, err := svc.List(context.Background(), 20, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "community hello", list[0].Body)
	require.Nil(t, list[0].GroupID)
}

func TestGroupPostRequiresActiveMember(t *testing.T) {
	svc, _, _ := newFeedSvc()
	_, err := svc.CreateForGroup(context.Background(), 4, 1, "stranger")
	require.ErrorIs(t, err, feed.ErrPostNotFound)

	_, err = svc.CreateForGroup(context.Background(), 3, 1, "pending")
	require.ErrorIs(t, err, feed.ErrPostNotFound)

	post, err := svc.CreateForGroup(context.Background(), 2, 1, "from member")
	require.NoError(t, err)
	require.NotNil(t, post.GroupID)
	require.Equal(t, int64(1), *post.GroupID)

	list, err := svc.ListForGroup(context.Background(), 2, 1, 20, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)

	_, err = svc.ListForGroup(context.Background(), 4, 1, 20, 0)
	require.ErrorIs(t, err, feed.ErrPostNotFound)
}

func TestOwnerCanDeleteMemberGroupPost(t *testing.T) {
	svc, posts, _ := newFeedSvc()
	post, err := svc.CreateForGroup(context.Background(), 2, 1, "member post")
	require.NoError(t, err)

	require.ErrorIs(t, svc.Delete(context.Background(), post.ID, 4), feed.ErrPostNotFound)
	require.NoError(t, svc.Delete(context.Background(), post.ID, 1))
	_, err = posts.GetByID(context.Background(), post.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestNonMemberCannotReactToGroupPost(t *testing.T) {
	svc, _, _ := newFeedSvc()
	post, err := svc.CreateForGroup(context.Background(), 1, 1, "tee time Saturday")
	require.NoError(t, err)

	_, err = svc.ToggleReaction(context.Background(), post.ID, 4, "golf")
	require.ErrorIs(t, err, feed.ErrPostNotFound)

	rx, err := svc.ToggleReaction(context.Background(), post.ID, 2, "golf")
	require.NoError(t, err)
	require.Len(t, rx, 1)
}
