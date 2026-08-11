package httphandler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httphandler "github.com/ericrabun/findfore-go/internal/adapter/inbound/http"
	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/application/friends"
	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

const testJWTSecret = "phase4-test-secret"

type stubTokenVersions struct {
	versions map[int64]int32
}

func (s stubTokenVersions) GetTokenVersion(_ context.Context, playerID int64) (int32, error) {
	if s.versions == nil {
		return 0, nil
	}
	v, ok := s.versions[playerID]
	if !ok {
		return 0, errors.New("missing player")
	}
	return v, nil
}

type stubPosts struct{}

func (stubPosts) List(context.Context, int32, int32) ([]entity.PostWithDetails, error) {
	return []entity.PostWithDetails{}, nil
}
func (stubPosts) Create(context.Context, int64, string) (*entity.PostWithDetails, error) {
	return nil, nil
}
func (stubPosts) Delete(context.Context, int64, int64) error { return nil }
func (stubPosts) ToggleReaction(context.Context, int64, int64, string) ([]entity.Reaction, error) {
	return nil, nil
}
func (stubPosts) CreateReply(context.Context, int64, int64, string) (*entity.Reply, error) {
	return nil, nil
}
func (stubPosts) DeleteReply(context.Context, int64, int64) error { return nil }

type stubEvents struct{}

func (stubEvents) List(context.Context, int64, *int64, bool) ([]entity.EventWithDetails, error) {
	return nil, nil
}
func (stubEvents) Get(context.Context, int64, int64) (*entity.EventWithDetails, error) {
	return nil, events.ErrEventNotFound
}
func (stubEvents) Create(context.Context, entity.Event, []int64) (*entity.EventWithDetails, error) {
	return nil, nil
}
func (stubEvents) Update(_ context.Context, actorID int64, e entity.Event, _ []int64) (*entity.EventWithDetails, error) {
	if actorID != 1 {
		return nil, events.ErrEventForbidden
	}
	return &entity.EventWithDetails{ID: e.ID, HostID: 1, CourseName: "Test", Date: e.Date, TeeTime: e.TeeTime, OpenSpots: e.OpenSpots, NumberOfHoles: e.NumberOfHoles}, nil
}
func (stubEvents) Delete(_ context.Context, actorID, _ int64) error {
	if actorID != 1 {
		return events.ErrEventForbidden
	}
	return nil
}
func (stubEvents) ListFriendsEvents(context.Context, int64) ([]entity.EventWithDetails, error) {
	return nil, nil
}

type stubPlayers struct{}

func (stubPlayers) List(context.Context) ([]entity.PlayerWithDetails, error) { return nil, nil }
func (stubPlayers) GetWithDetails(context.Context, int64) (*entity.PlayerWithDetails, error) {
	return nil, nil
}
func (stubPlayers) Create(context.Context, string, string, string, string, string, string) (*entity.Player, error) {
	return nil, nil
}
func (stubPlayers) Update(context.Context, int64, string, string, string, string) (*entity.PlayerWithDetails, error) {
	return nil, nil
}
func (stubPlayers) ChangePassword(context.Context, int64, string, string, string) error { return nil }

type stubSessions struct{}

func (stubSessions) Login(context.Context, string, string) (*entity.PlayerWithDetails, string, error) {
	return nil, "", errors.New("unused")
}

type stubCourses struct{}

func (stubCourses) List(context.Context) ([]entity.Course, error) { return nil, nil }
func (stubCourses) Search(context.Context, string) ([]entity.Course, error) {
	return nil, nil
}
func (stubCourses) FindOrCreate(context.Context, entity.Course) (*entity.Course, bool, error) {
	return nil, false, nil
}

type stubPlayerEvents struct {
	joinErr   error
	updateErr error
}

func (s stubPlayerEvents) UpdateStatus(context.Context, int64, int64, string) (*entity.PlayerEvent, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	return &entity.PlayerEvent{ID: 1, PlayerID: 1, EventID: 10, InviteStatus: entity.InviteStatusAccepted}, nil
}
func (s stubPlayerEvents) JoinEvent(context.Context, int64, int64) (*entity.PlayerEvent, error) {
	if s.joinErr != nil {
		return nil, s.joinErr
	}
	return &entity.PlayerEvent{ID: 1, PlayerID: 1, EventID: 10, InviteStatus: entity.InviteStatusAccepted}, nil
}

type stubFriendships struct {
	acceptErr error
}

func (s stubFriendships) Request(context.Context, int32, int32) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error) {
	return nil, nil, nil, nil
}
func (s stubFriendships) Accept(context.Context, int32, int64) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error) {
	if s.acceptErr != nil {
		return nil, nil, nil, s.acceptErr
	}
	return &entity.Friendship{ID: 1, Status: entity.FriendshipStatusAccepted}, &entity.PlayerWithDetails{ID: 1}, &entity.PlayerWithDetails{ID: 2}, nil
}
func (s stubFriendships) Decline(context.Context, int32, int64) error { return nil }
func (s stubFriendships) CancelOrUnfriend(context.Context, int32, int64) error {
	return nil
}
func (s stubFriendships) ListIncomingRequests(context.Context, int32) ([]entity.Friendship, error) {
	return nil, nil
}
func (s stubFriendships) ListOutgoingPendingIDs(context.Context, int32) ([]int64, error) {
	return nil, nil
}
func (s stubFriendships) ListAccepted(context.Context, int32) ([]entity.Friendship, error) {
	return nil, nil
}

func testRouter(friendships stubFriendships) http.Handler {
	return testRouterWith(stubPlayerEvents{}, friendships)
}

func testRouterWith(playerEvents stubPlayerEvents, friendships stubFriendships) http.Handler {
	h := httphandler.New(stubPlayers{}, stubSessions{}, stubCourses{}, stubEvents{}, playerEvents, friendships, stubPosts{})
	return httphandler.NewRouter(h, testJWTSecret, stubTokenVersions{versions: map[int64]int32{1: 0, 2: 0}})
}

func bearer(playerID int64) string {
	token, err := auth.GenerateToken(playerID, 0, testJWTSecret)
	if err != nil {
		panic(err)
	}
	return "Bearer " + token
}

func TestProtectedRoutesRequireAuth(t *testing.T) {
	r := testRouter(stubFriendships{})

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/posts"},
		{http.MethodGet, "/api/v1/players"},
		{http.MethodGet, "/api/v1/events"},
		{http.MethodDelete, "/api/v1/event/1"},
	}

	for _, tc := range paths {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

func TestPostsListWithAuth(t *testing.T) {
	r := testRouter(stubFriendships{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	req.Header.Set("Authorization", bearer(1))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteEventForbiddenForNonHost(t *testing.T) {
	r := testRouter(stubFriendships{})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/event/5", nil)
	req.Header.Set("Authorization", bearer(2))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	errorsList, _ := body["errors"].([]any)
	require.NotEmpty(t, errorsList)
}

func TestAcceptFriendshipForbiddenMapsTo403(t *testing.T) {
	r := testRouter(stubFriendships{acceptErr: friends.ErrFriendshipForbidden})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/friendships/9/accept", nil)
	req.Header.Set("Authorization", bearer(1))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestJoinEventConflictMapsTo409(t *testing.T) {
	r := testRouterWith(stubPlayerEvents{joinErr: entity.ErrEventFull}, stubFriendships{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/player-event/join", strings.NewReader(`{"event_id":10}`))
	req.Header.Set("Authorization", bearer(1))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAcceptInviteConflictMapsTo409(t *testing.T) {
	r := testRouterWith(stubPlayerEvents{updateErr: entity.ErrEventFull}, stubFriendships{})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/player-event", strings.NewReader(`{"event_id":10,"invite_status":"accepted"}`))
	req.Header.Set("Authorization", bearer(1))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)
}
