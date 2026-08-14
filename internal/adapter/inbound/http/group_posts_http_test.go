package httphandler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	httphandler "github.com/ericrabun/findfore-go/internal/adapter/inbound/http"
	"github.com/ericrabun/findfore-go/internal/application/feed"
	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type httpFakePosts struct {
	byID   map[int64]*entity.PostWithDetails
	nextID int64
}

func newHTTPFakePosts() *httpFakePosts {
	return &httpFakePosts{byID: map[int64]*entity.PostWithDetails{}, nextID: 1}
}

func (f *httpFakePosts) GetByID(_ context.Context, id int64) (*entity.PostWithDetails, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}
func (f *httpFakePosts) List(_ context.Context, _, _ int32) ([]entity.PostWithDetails, error) {
	var out []entity.PostWithDetails
	for _, p := range f.byID {
		if p.GroupID == nil {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (f *httpFakePosts) ListByGroupID(_ context.Context, groupID int64, _, _ int32) ([]entity.PostWithDetails, error) {
	var out []entity.PostWithDetails
	for _, p := range f.byID {
		if p.GroupID != nil && *p.GroupID == groupID {
			out = append(out, *p)
		}
	}
	return out, nil
}
func (f *httpFakePosts) Create(_ context.Context, playerID int64, body string, groupID *int64) (int64, error) {
	id := f.nextID
	f.nextID++
	f.byID[id] = &entity.PostWithDetails{
		ID: id, PlayerID: playerID, PlayerName: "Player", Body: body,
		GroupID: groupID, CreatedAt: time.Now().UTC(),
	}
	return id, nil
}
func (f *httpFakePosts) Delete(_ context.Context, id, playerID int64) error {
	p, ok := f.byID[id]
	if !ok || p.PlayerID != playerID {
		return sql.ErrNoRows
	}
	delete(f.byID, id)
	return nil
}
func (f *httpFakePosts) DeleteByID(_ context.Context, id int64) error {
	delete(f.byID, id)
	return nil
}

type httpFakeReactions struct{}

func (httpFakeReactions) ListByPostID(context.Context, int64) ([]entity.Reaction, error) {
	return nil, nil
}
func (httpFakeReactions) Find(context.Context, int64, int64, string) (*entity.Reaction, error) {
	return nil, sql.ErrNoRows
}
func (httpFakeReactions) Create(context.Context, int64, int64, string) (*entity.Reaction, error) {
	return &entity.Reaction{ID: 1}, nil
}
func (httpFakeReactions) Delete(context.Context, int64, int64, string) error { return nil }

type httpFakeReplies struct{}

func (httpFakeReplies) ListByPostID(context.Context, int64) ([]entity.Reply, error) { return nil, nil }
func (httpFakeReplies) Create(context.Context, int64, int64, string) (*entity.Reply, error) {
	return &entity.Reply{ID: 1}, nil
}
func (httpFakeReplies) Delete(context.Context, int64, int64) error { return nil }

func newGroupPostsHTTPEnv(t *testing.T) *groupsHTTPEnv {
	t.Helper()
	repo := newHTTPFakeGroups()
	posts := newHTTPFakePosts()
	groupSvc := groups.NewService(repo, httpFakePlayers{})
	postSvc := feed.NewService(posts, httpFakeReactions{}, httpFakeReplies{}, repo)
	h := httphandler.New(stubPlayers{}, stubSessions{}, stubCourses{}, stubEvents{}, stubPlayerEvents{}, stubFriendships{}, postSvc, nil, groupSvc)
	return &groupsHTTPEnv{
		router: httphandler.NewRouter(h, testJWTSecret, stubTokenVersions{versions: map[int64]int32{1: 0, 2: 0, 3: 0}}),
		repo:   repo,
	}
}

func TestGroupPostsHTTPMemberCreateAndList(t *testing.T) {
	e := newGroupPostsHTTPEnv(t)
	rec := e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)

	rec = e.do(t, http.MethodPost, "/api/v1/groups/1/posts", `{"body":"Saturday 8am?"}`, 2)
	require.Equal(t, http.StatusCreated, rec.Code)

	rec = e.do(t, http.MethodGet, "/api/v1/groups/1/posts", "", 2)
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Posts []struct {
			Body string `json:"body"`
		} `json:"posts"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Posts, 1)
	require.Equal(t, "Saturday 8am?", body.Posts[0].Body)
}

func TestGroupPostsHTTPNonMemberHidden(t *testing.T) {
	e := newGroupPostsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	e.do(t, http.MethodPost, "/api/v1/groups/1/posts", `{"body":"members only"}`, 1)

	rec := e.do(t, http.MethodGet, "/api/v1/groups/1/posts", "", 2)
	require.Equal(t, http.StatusNotFound, rec.Code)
	rec = e.do(t, http.MethodPost, "/api/v1/groups/1/posts", `{"body":"nope"}`, 2)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGroupPostsHTTPBlankBody(t *testing.T) {
	e := newGroupPostsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	rec := e.do(t, http.MethodPost, "/api/v1/groups/1/posts", `{"body":""}`, 1)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
