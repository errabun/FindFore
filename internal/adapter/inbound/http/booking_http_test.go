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
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/fakebooking"
	"github.com/ericrabun/findfore-go/internal/application/booking"
	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
	"github.com/google/uuid"
)

// --- in-memory booking repos for HTTP tests ---

type httpFakeTees struct {
	byID       map[int64]*entity.TeeTime
	byProvider map[string]*entity.TeeTime
	providers  map[string]*entity.TeeTimeProvider
	byTeeProv  map[string]*entity.TeeTimeProvider
	nextID     int64
}

func newHTTPFakeTees() *httpFakeTees {
	return &httpFakeTees{
		byID: map[int64]*entity.TeeTime{}, byProvider: map[string]*entity.TeeTime{},
		providers: map[string]*entity.TeeTimeProvider{}, byTeeProv: map[string]*entity.TeeTimeProvider{},
		nextID: 1,
	}
}

func pk(provider, externalID string) string { return provider + ":" + externalID }
func tk(id int64, provider string) string   { return fmt.Sprintf("%s/%d", provider, id) }

func (f *httpFakeTees) GetByID(_ context.Context, id int64) (*entity.TeeTime, error) {
	t, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *t
	return &cp, nil
}
func (f *httpFakeTees) ListByCourseAndWindow(_ context.Context, courseID int64, from, to time.Time) ([]entity.TeeTime, error) {
	var out []entity.TeeTime
	for _, t := range f.byID {
		if t.CourseID == courseID && !t.StartsAt.Before(from) && t.StartsAt.Before(to) {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (f *httpFakeTees) GetByProviderExternalID(_ context.Context, provider, externalID string) (*entity.TeeTime, error) {
	t, ok := f.byProvider[pk(provider, externalID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *t
	return &cp, nil
}
func (f *httpFakeTees) Create(_ context.Context, t entity.TeeTime) (*entity.TeeTime, error) {
	t.ID = f.nextID
	f.nextID++
	cp := t
	f.byID[t.ID] = &cp
	return &t, nil
}
func (f *httpFakeTees) UpdateCache(_ context.Context, t entity.TeeTime) (*entity.TeeTime, error) {
	cp := t
	f.byID[t.ID] = &cp
	for k, v := range f.byProvider {
		if v.ID == t.ID {
			f.byProvider[k] = &cp
		}
	}
	return &t, nil
}
func (f *httpFakeTees) UpdateStatus(_ context.Context, id int64, status string) (*entity.TeeTime, error) {
	t := f.byID[id]
	t.Status = status
	cp := *t
	return &cp, nil
}
func (f *httpFakeTees) GetProvider(_ context.Context, provider, externalID string) (*entity.TeeTimeProvider, error) {
	p, ok := f.providers[pk(provider, externalID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}
func (f *httpFakeTees) GetProviderByTeeTime(_ context.Context, teeTimeID int64, provider string) (*entity.TeeTimeProvider, error) {
	p, ok := f.byTeeProv[tk(teeTimeID, provider)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}
func (f *httpFakeTees) LinkProvider(_ context.Context, teeTimeID int64, provider, externalID string) error {
	p := &entity.TeeTimeProvider{ID: int64(len(f.providers) + 1), TeeTimeID: teeTimeID, Provider: provider, ExternalID: externalID}
	f.providers[pk(provider, externalID)] = p
	f.byTeeProv[tk(teeTimeID, provider)] = p
	if t, ok := f.byID[teeTimeID]; ok {
		f.byProvider[pk(provider, externalID)] = t
	}
	return nil
}

type httpFakeRes struct {
	byID    map[int64]*entity.Reservation
	active  map[int64]int64
	players map[int64][]entity.ReservationPlayer
	nextID  int64
}

func newHTTPFakeRes() *httpFakeRes {
	return &httpFakeRes{
		byID: map[int64]*entity.Reservation{}, active: map[int64]int64{},
		players: map[int64][]entity.ReservationPlayer{}, nextID: 1,
	}
}

func (f *httpFakeRes) GetByID(_ context.Context, id int64) (*entity.Reservation, error) {
	r, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *r
	return &cp, nil
}
func (f *httpFakeRes) GetActiveByTeeTimeID(_ context.Context, teeTimeID int64) (*entity.Reservation, error) {
	id, ok := f.active[teeTimeID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return f.GetByID(context.Background(), id)
}
func (f *httpFakeRes) Create(_ context.Context, r entity.Reservation, players []entity.ReservationPlayer) (*entity.Reservation, error) {
	if _, ok := f.active[r.TeeTimeID]; ok && entity.IsActiveReservation(r.Status) {
		return nil, entity.ErrActiveReservationExists
	}
	if r.ProviderRequestID == "" {
		r.ProviderRequestID = uuid.NewString()
	}
	r.ID = f.nextID
	f.nextID++
	now := time.Now().UTC()
	r.CreatedAt, r.UpdatedAt = now, now
	cp := r
	f.byID[r.ID] = &cp
	if entity.IsActiveReservation(r.Status) {
		f.active[r.TeeTimeID] = r.ID
	}
	f.players[r.ID] = append([]entity.ReservationPlayer(nil), players...)
	return &r, nil
}
func (f *httpFakeRes) Update(_ context.Context, r entity.Reservation) (*entity.Reservation, error) {
	cp := r
	cp.UpdatedAt = time.Now().UTC()
	f.byID[r.ID] = &cp
	if entity.IsActiveReservation(r.Status) {
		f.active[r.TeeTimeID] = r.ID
	} else if f.active[r.TeeTimeID] == r.ID {
		delete(f.active, r.TeeTimeID)
	}
	return &cp, nil
}
func (f *httpFakeRes) ListPlayers(_ context.Context, reservationID int64) ([]entity.ReservationPlayer, error) {
	return append([]entity.ReservationPlayer(nil), f.players[reservationID]...), nil
}

type httpFakeCourses struct {
	providers map[string]*entity.CourseProvider
}

func newHTTPFakeCourses() *httpFakeCourses {
	return &httpFakeCourses{providers: map[string]*entity.CourseProvider{}}
}

func (f *httpFakeCourses) List(context.Context) ([]entity.Course, error)          { return nil, nil }
func (f *httpFakeCourses) GetByID(context.Context, int64) (*entity.Course, error) { return nil, sql.ErrNoRows }
func (f *httpFakeCourses) GetByNameAndCity(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (f *httpFakeCourses) GetByProviderExternalID(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (f *httpFakeCourses) Create(context.Context, entity.Course) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (f *httpFakeCourses) GetProvider(context.Context, string, string) (*entity.CourseProvider, error) {
	return nil, sql.ErrNoRows
}
func (f *httpFakeCourses) GetProviderByCourse(_ context.Context, courseID int64, provider string) (*entity.CourseProvider, error) {
	p, ok := f.providers[tk(courseID, provider)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}
func (f *httpFakeCourses) LinkProvider(_ context.Context, courseID int64, provider, externalID string) error {
	f.providers[tk(courseID, provider)] = &entity.CourseProvider{CourseID: courseID, Provider: provider, ExternalID: externalID}
	return nil
}

type bookingHTTPEnv struct {
	router   http.Handler
	provider *fakebooking.Provider
	tees     *httpFakeTees
	res      *httpFakeRes
	courses  *httpFakeCourses
}

func newBookingHTTPEnv(t *testing.T) *bookingHTTPEnv {
	t.Helper()
	tees := newHTTPFakeTees()
	res := newHTTPFakeRes()
	courses := newHTTPFakeCourses()
	require.NoError(t, courses.LinkProvider(context.Background(), 1, entity.ProviderLightspeed, "course-ext"))
	provider := fakebooking.New(entity.ProviderLightspeed)
	svc := booking.NewService(tees, res, courses, provider)
	h := httphandler.New(stubPlayers{}, stubSessions{}, stubCourses{}, stubEvents{}, stubPlayerEvents{}, stubFriendships{}, stubPosts{}, svc)
	return &bookingHTTPEnv{
		router: httphandler.NewRouter(h, testJWTSecret, stubTokenVersions{versions: map[int64]int32{1: 0, 2: 0}}),
		provider: provider, tees: tees, res: res, courses: courses,
	}
}

func (e *bookingHTTPEnv) do(t *testing.T, method, path, body string, playerID int64) *httptest.ResponseRecorder {
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

func assertNoProviderLeak(t *testing.T, body []byte) {
	t.Helper()
	s := string(body)
	require.NotContains(t, s, `"provider"`)
	require.NotContains(t, s, "provider_request_id")
	require.NotContains(t, s, "external_reservation")
	require.NotContains(t, s, "external_id")
	require.NotContains(t, s, "lightspeed")
}

func TestBookingHTTPUnauthenticated(t *testing.T) {
	e := newBookingHTTPEnv(t)
	rec := e.do(t, http.MethodGet, "/api/v1/courses/1/tee-times?from=2026-08-15T13:00:00Z&to=2026-08-15T20:00:00Z", "", 0)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestBookingHTTPHappyPath(t *testing.T) {
	e := newBookingHTTPEnv(t)
	from := time.Date(2026, 8, 15, 13, 0, 0, 0, time.UTC)
	to := from.Add(8 * time.Hour)
	price, slots := int32(6500), int32(4)
	e.provider.Slots = []port.BookingSlot{{
		ExternalID: "ls-1", StartsAt: from.Add(time.Hour), Holes: "18",
		AvailableSlots: &slots, PriceCents: &price, Currency: "USD",
		Status: entity.TeeTimeStatusAvailable,
	}}

	rec := e.do(t, http.MethodGet,
		fmt.Sprintf("/api/v1/courses/1/tee-times?from=%s&to=%s&players=2", from.Format(time.RFC3339), to.Format(time.RFC3339)),
		"", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	assertNoProviderLeak(t, rec.Body.Bytes())
	var list struct {
		TeeTimes []struct {
			ID int64 `json:"id"`
		} `json:"tee_times"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list.TeeTimes, 1)
	teeID := list.TeeTimes[0].ID

	body := fmt.Sprintf(`{"tee_time_id":%d,"players":[{"player_id":1},{"guest_name":"John"}]}`, teeID)
	rec = e.do(t, http.MethodPost, "/api/v1/reservations", body, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	assertNoProviderLeak(t, rec.Body.Bytes())
	var res struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.Equal(t, "held", res.Status)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	assertNoProviderLeak(t, rec.Body.Bytes())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.Equal(t, "confirmed", res.Status)

	// Idempotent confirm
	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.Equal(t, "cancelled", res.Status)

	// Idempotent cancel
	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestBookingHTTPTimeoutThenRetry(t *testing.T) {
	e := newBookingHTTPEnv(t)
	price := int32(6500)
	tt, err := e.tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, PriceCents: &price, Currency: "USD",
	})
	require.NoError(t, err)
	require.NoError(t, e.tees.LinkProvider(context.Background(), tt.ID, entity.ProviderLightspeed, "ls-slot"))

	e.provider.HoldBehavior = fakebooking.BehaviorTimeout
	body := fmt.Sprintf(`{"tee_time_id":%d,"players":[{"player_id":1}]}`, tt.ID)
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", body, 1)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	e.provider.HoldBehavior = fakebooking.BehaviorSuccess
	rec = e.do(t, http.MethodPost, "/api/v1/reservations", body, 1)
	require.Equal(t, http.StatusOK, rec.Code) // resume, not created
	var res struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.Equal(t, "held", res.Status)
	assertNoProviderLeak(t, rec.Body.Bytes())
}

func TestBookingHTTPProviderReject(t *testing.T) {
	e := newBookingHTTPEnv(t)
	price := int32(6500)
	tt, err := e.tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, PriceCents: &price, Currency: "USD",
	})
	require.NoError(t, err)
	require.NoError(t, e.tees.LinkProvider(context.Background(), tt.ID, entity.ProviderLightspeed, "ls-slot"))
	e.provider.HoldBehavior = fakebooking.BehaviorReject

	body := fmt.Sprintf(`{"tee_time_id":%d,"players":[{"player_id":1}]}`, tt.ID)
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", body, 1)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestBookingHTTPStaleInventory(t *testing.T) {
	e := newBookingHTTPEnv(t)
	from := time.Date(2026, 8, 15, 13, 0, 0, 0, time.UTC)
	to := from.Add(8 * time.Hour)
	price, slots := int32(6500), int32(4)
	e.provider.Slots = []port.BookingSlot{{
		ExternalID: "ls-stale", StartsAt: from.Add(time.Hour),
		AvailableSlots: &slots, PriceCents: &price, Currency: "USD",
		Status: entity.TeeTimeStatusAvailable,
	}}
	rec := e.do(t, http.MethodGet,
		fmt.Sprintf("/api/v1/courses/1/tee-times?from=%s&to=%s", from.Format(time.RFC3339), to.Format(time.RFC3339)),
		"", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	var list struct {
		TeeTimes []struct {
			ID int64 `json:"id"`
		} `json:"tee_times"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list.TeeTimes, 1)

	e.provider.GoneExternalIDs["ls-stale"] = true
	body := fmt.Sprintf(`{"tee_time_id":%d,"players":[{"player_id":1}]}`, list.TeeTimes[0].ID)
	rec = e.do(t, http.MethodPost, "/api/v1/reservations", body, 1)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestBookingHTTPCancelFailure(t *testing.T) {
	e := newBookingHTTPEnv(t)
	price := int32(6500)
	tt, err := e.tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, PriceCents: &price, Currency: "USD",
	})
	require.NoError(t, err)
	require.NoError(t, e.tees.LinkProvider(context.Background(), tt.ID, entity.ProviderLightspeed, "ls-slot"))

	body := fmt.Sprintf(`{"tee_time_id":%d,"players":[{"player_id":1}]}`, tt.ID)
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", body, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	var res struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)

	e.provider.CancelBehavior = fakebooking.BehaviorCancelFail
	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestBookingHTTPAuthZ(t *testing.T) {
	e := newBookingHTTPEnv(t)
	price := int32(6500)
	tt, err := e.tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, PriceCents: &price, Currency: "USD",
	})
	require.NoError(t, err)
	require.NoError(t, e.tees.LinkProvider(context.Background(), tt.ID, entity.ProviderLightspeed, "ls-slot"))

	body := fmt.Sprintf(`{"tee_time_id":%d,"players":[{"player_id":1}]}`, tt.ID)
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", body, 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	var res struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 2)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestBookingHTTPRejectsProviderFieldsInBody(t *testing.T) {
	// Server ignores unknown fields; ensure we never echo them.
	e := newBookingHTTPEnv(t)
	price := int32(6500)
	tt, err := e.tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, PriceCents: &price, Currency: "USD",
	})
	require.NoError(t, err)
	require.NoError(t, e.tees.LinkProvider(context.Background(), tt.ID, entity.ProviderLightspeed, "ls-slot"))

	payload := map[string]any{
		"tee_time_id":         tt.ID,
		"provider":            "lightspeed",
		"provider_request_id": "client-should-not-set",
		"players":             []map[string]any{{"player_id": 1}},
	}
	b, _ := json.Marshal(payload)
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", string(b), 1)
	require.Equal(t, http.StatusCreated, rec.Code)
	assertNoProviderLeak(t, rec.Body.Bytes())
}
