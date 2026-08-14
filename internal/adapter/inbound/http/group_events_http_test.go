package httphandler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	httphandler "github.com/ericrabun/findfore-go/internal/adapter/inbound/http"
	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type httpFakeEvents struct {
	byID    map[int64]*entity.Event
	details map[int64]*entity.EventWithDetails
	nextID  int64
}

func newHTTPFakeEvents() *httpFakeEvents {
	return &httpFakeEvents{
		byID:    map[int64]*entity.Event{},
		details: map[int64]*entity.EventWithDetails{},
		nextID:  1,
	}
}

func (f *httpFakeEvents) GetByID(_ context.Context, id int64) (*entity.Event, error) {
	e, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *e
	return &cp, nil
}
func (f *httpFakeEvents) GetDetailsByID(_ context.Context, id int64) (*entity.EventWithDetails, error) {
	d, ok := f.details[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *d
	return &cp, nil
}
func (f *httpFakeEvents) ListAllIDs(context.Context) ([]int64, error) { return nil, nil }
func (f *httpFakeEvents) ListPublicIDs(context.Context) ([]int64, error) {
	var ids []int64
	for id, d := range f.details {
		if !d.Private {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
func (f *httpFakeEvents) ListIDsByPlayerID(context.Context, int64) ([]int64, error) { return nil, nil }
func (f *httpFakeEvents) ListFriendsAvailableIDs(context.Context, int32, int64) ([]int64, error) {
	return nil, nil
}
func (f *httpFakeEvents) ListIDsByGroupID(_ context.Context, groupID int64) ([]int64, error) {
	var ids []int64
	for id, e := range f.byID {
		if e.GroupID != nil && *e.GroupID == groupID {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
func (f *httpFakeEvents) Create(context.Context, entity.Event) (int64, error) { return 0, nil }
func (f *httpFakeEvents) CreateWithInvites(_ context.Context, e entity.Event, _ []int64) (int64, error) {
	e.ID = f.nextID
	f.nextID++
	cp := e
	f.byID[e.ID] = &cp
	f.details[e.ID] = &entity.EventWithDetails{
		ID: e.ID, CourseName: "Test Course", CourseTimezone: entity.DefaultCourseTimezone,
		PlannedStartsAt: e.PlannedStartsAt, GroupID: e.GroupID, OpenSpots: e.OpenSpots,
		NumberOfHoles: e.NumberOfHoles, Private: e.Private, HostID: e.HostID, HostName: "Host",
	}
	return e.ID, nil
}
func (f *httpFakeEvents) Update(context.Context, entity.Event) error { return nil }
func (f *httpFakeEvents) Delete(context.Context, int64) error        { return nil }
func (f *httpFakeEvents) DeletePast(context.Context) error           { return nil }

type httpFakePlayerEvents struct {
	accepted map[int64][]int64
}

func newHTTPFakePlayerEvents() *httpFakePlayerEvents {
	return &httpFakePlayerEvents{accepted: map[int64][]int64{}}
}

func (f *httpFakePlayerEvents) Get(context.Context, int64, int64) (*entity.PlayerEvent, error) {
	return nil, sql.ErrNoRows
}
func (f *httpFakePlayerEvents) Create(context.Context, entity.PlayerEvent) (*entity.PlayerEvent, error) {
	return nil, nil
}
func (f *httpFakePlayerEvents) UpdateStatus(context.Context, int64, int64, entity.InviteStatus) (*entity.PlayerEvent, error) {
	return nil, nil
}
func (f *httpFakePlayerEvents) ListPlayerIDsByEventAndStatus(_ context.Context, eventID int64, status entity.InviteStatus) ([]int64, error) {
	if status == entity.InviteStatusAccepted {
		return f.accepted[eventID], nil
	}
	return nil, nil
}
func (f *httpFakePlayerEvents) CountAcceptedForEvent(context.Context, int64) (int64, error) {
	return 0, nil
}
func (f *httpFakePlayerEvents) ClosePendingForEvent(context.Context, int64) error { return nil }
func (f *httpFakePlayerEvents) ReopenClosedForEvent(context.Context, int64) error { return nil }
func (f *httpFakePlayerEvents) JoinAccepted(context.Context, int64, int64) (*entity.PlayerEvent, error) {
	return nil, entity.ErrEventMissing
}
func (f *httpFakePlayerEvents) AcceptInvite(context.Context, int64, int64) (*entity.PlayerEvent, error) {
	return nil, entity.ErrPlayerEventMissing
}

type httpFakeEventCourses struct{}

func (httpFakeEventCourses) List(context.Context) ([]entity.Course, error) { return nil, nil }
func (httpFakeEventCourses) GetByID(_ context.Context, id int64) (*entity.Course, error) {
	return &entity.Course{ID: id, Timezone: entity.DefaultCourseTimezone}, nil
}
func (httpFakeEventCourses) GetByNameAndCity(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (httpFakeEventCourses) GetByProviderExternalID(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (httpFakeEventCourses) Create(context.Context, entity.Course) (*entity.Course, error) {
	return nil, nil
}
func (httpFakeEventCourses) GetProvider(context.Context, string, string) (*entity.CourseProvider, error) {
	return nil, sql.ErrNoRows
}
func (httpFakeEventCourses) GetProviderByCourse(context.Context, int64, string) (*entity.CourseProvider, error) {
	return nil, sql.ErrNoRows
}
func (httpFakeEventCourses) LinkProvider(context.Context, int64, string, string) error { return nil }

func newGroupEventsHTTPEnv(t *testing.T) *groupsHTTPEnv {
	t.Helper()
	repo := newHTTPFakeGroups()
	eventRepo := newHTTPFakeEvents()
	playerEvents := newHTTPFakePlayerEvents()
	groupSvc := groups.NewService(repo, httpFakePlayers{})
	eventSvc := events.NewService(eventRepo, playerEvents, httpFakeEventCourses{}, repo)
	h := httphandler.New(stubPlayers{}, stubSessions{}, stubCourses{}, eventSvc, stubPlayerEvents{}, stubFriendships{}, stubPosts{}, nil, groupSvc)
	return &groupsHTTPEnv{
		router: httphandler.NewRouter(h, testJWTSecret, stubTokenVersions{versions: map[int64]int32{1: 0, 2: 0, 3: 0}}),
		repo:   repo,
	}
}

func TestGroupEventsHTTPMemberCreateAndList(t *testing.T) {
	e := newGroupEventsHTTPEnv(t)
	rec := e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	e.do(t, http.MethodPost, "/api/v1/groups/1/join", "", 2)

	body := `{"course_id":1,"date":"2099-01-01","tee_time":"08:00","open_spots":4,"number_of_holes":"18","private":false,"group_id":1}`
	rec = e.do(t, http.MethodPost, "/api/v1/event", body, 2)
	require.Equal(t, http.StatusCreated, rec.Code)
	var created struct {
		Private bool   `json:"private"`
		GroupID *int64 `json:"group_id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.True(t, created.Private)
	require.NotNil(t, created.GroupID)
	require.Equal(t, int64(1), *created.GroupID)

	rec = e.do(t, http.MethodGet, "/api/v1/groups/1/events", "", 2)
	require.Equal(t, http.StatusOK, rec.Code)
	var listed struct {
		Events []struct {
			CourseName string `json:"course_name"`
		} `json:"events"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &listed))
	require.Len(t, listed.Events, 1)
	require.Equal(t, "Test Course", listed.Events[0].CourseName)
}

func TestGroupEventsHTTPNonMemberHidden(t *testing.T) {
	e := newGroupEventsHTTPEnv(t)
	e.do(t, http.MethodPost, "/api/v1/groups", `{"name":"Crew","privacy":"public"}`, 1)

	body := `{"course_id":1,"date":"2099-01-01","tee_time":"08:00","open_spots":4,"number_of_holes":"18","group_id":1}`
	rec := e.do(t, http.MethodPost, "/api/v1/event", body, 1)
	require.Equal(t, http.StatusCreated, rec.Code)

	rec = e.do(t, http.MethodGet, "/api/v1/groups/1/events", "", 2)
	require.Equal(t, http.StatusNotFound, rec.Code)
	rec = e.do(t, http.MethodPost, "/api/v1/event", body, 2)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
