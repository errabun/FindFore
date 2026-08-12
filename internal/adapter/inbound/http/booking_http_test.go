package httphandler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
func (f *httpFakeRes) GetByClientIdempotency(_ context.Context, bookedByPlayerID int64, clientIdempotencyKey string) (*entity.Reservation, error) {
	for _, r := range f.byID {
		if r.BookedByPlayerID == bookedByPlayerID && r.ClientIdempotencyKey == clientIdempotencyKey {
			cp := *r
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}
func (f *httpFakeRes) Create(_ context.Context, r entity.Reservation, players []entity.ReservationPlayer) (*entity.Reservation, error) {
	if _, ok := f.active[r.TeeTimeID]; ok && entity.IsActiveReservation(r.Status) {
		return nil, entity.ErrActiveReservationExists
	}
	if r.ClientIdempotencyKey != "" {
		for _, existing := range f.byID {
			if existing.BookedByPlayerID == r.BookedByPlayerID && existing.ClientIdempotencyKey == r.ClientIdempotencyKey {
				return nil, entity.ErrActiveReservationExists
			}
		}
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

type httpFakePlayers struct{}

func (httpFakePlayers) List(context.Context) ([]entity.Player, error) { return nil, nil }
func (httpFakePlayers) GetByID(_ context.Context, id int64) (*entity.Player, error) {
	if id <= 0 {
		return nil, sql.ErrNoRows
	}
	return &entity.Player{ID: id, Name: "Player"}, nil
}
func (httpFakePlayers) GetByEmail(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (httpFakePlayers) GetByUsername(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (httpFakePlayers) Create(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (httpFakePlayers) Update(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (httpFakePlayers) GetPasswordByID(context.Context, int64) (string, error) { return "", sql.ErrNoRows }
func (httpFakePlayers) UpdatePassword(context.Context, int64, string) error    { return sql.ErrNoRows }
func (httpFakePlayers) GetTokenVersion(context.Context, int64) (int32, error)  { return 0, nil }
func (httpFakePlayers) ListIDsExcept(context.Context, int64) ([]int64, error)  { return nil, nil }

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
	svc := booking.NewService(tees, res, courses, httpFakePlayers{}, provider)
	h := httphandler.New(stubPlayers{}, stubSessions{}, stubCourses{}, stubEvents{}, stubPlayerEvents{}, stubFriendships{}, stubPosts{}, svc)
	return &bookingHTTPEnv{
		router: httphandler.NewRouter(h, testJWTSecret, stubTokenVersions{versions: map[int64]int32{1: 0, 2: 0}}),
		provider: provider, tees: tees, res: res, courses: courses,
	}
}

func (e *bookingHTTPEnv) do(t *testing.T, method, path, body string, playerID int64, idemKey ...string) *httptest.ResponseRecorder {
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
	if method == http.MethodPost && path == "/api/v1/reservations" {
		key := uuid.NewString()
		if len(idemKey) > 0 && idemKey[0] != "" {
			key = idemKey[0]
		}
		r.Header.Set("Idempotency-Key", key)
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, r)
	return rec
}

func (e *bookingHTTPEnv) seedLinkedTee(t *testing.T, externalID string) *entity.TeeTime {
	t.Helper()
	price, slots := int32(6500), int32(4)
	tt, err := e.tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, PriceCents: &price, Currency: "USD",
		AvailableSlots: &slots,
	})
	require.NoError(t, err)
	require.NoError(t, e.tees.LinkProvider(context.Background(), tt.ID, entity.ProviderLightspeed, externalID))
	return tt
}

func beginBody(teeID int64, playersJSON string) string {
	return fmt.Sprintf(`{"tee_time_id":%d,"players":%s}`, teeID, playersJSON)
}

type reservationJSON struct {
	ID        int64  `json:"id"`
	TeeTimeID int64  `json:"tee_time_id"`
	Status    string `json:"status"`
	PartySize int32  `json:"party_size"`
	Players   []struct {
		PlayerID  *int64 `json:"player_id"`
		GuestName string `json:"guest_name"`
	} `json:"players"`
}

func decodeReservation(t *testing.T, rec *httptest.ResponseRecorder) reservationJSON {
	t.Helper()
	var res reservationJSON
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return res
}

func assertNoProviderLeak(t *testing.T, body []byte) {
	t.Helper()
	s := string(body)
	require.NotContains(t, s, `"provider":`)
	require.NotContains(t, s, "provider_request_id")
	require.NotContains(t, s, "external_reservation")
	require.NotContains(t, s, "external_id")
	require.NotContains(t, s, "lightspeed")
	require.NotContains(t, s, "client_idempotency")
}

func TestBookingHTTPUnauthenticated(t *testing.T) {
	e := newBookingHTTPEnv(t)
	rec := e.do(t, http.MethodGet, "/api/v1/courses/1/tee-times?from=2026-08-15T13:00:00Z&to=2026-08-15T20:00:00Z", "", 0)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, 0, e.provider.SearchCalls)
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
	require.Equal(t, 1, e.provider.SearchCalls)
	var list struct {
		TeeTimes []struct {
			ID int64 `json:"id"`
		} `json:"tee_times"`
		Source    string `json:"source"`
		FetchedAt string `json:"fetched_at"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Equal(t, "provider", list.Source)
	require.NotEmpty(t, list.FetchedAt)
	require.Len(t, list.TeeTimes, 1)
	teeID := list.TeeTimes[0].ID

	body := beginBody(teeID, `[{"player_id":1},{"guest_name":"John"}]`)
	rec = e.do(t, http.MethodPost, "/api/v1/reservations", body, 1, "happy-key")
	require.Equal(t, http.StatusCreated, rec.Code)
	assertNoProviderLeak(t, rec.Body.Bytes())
	res := decodeReservation(t, rec)
	require.Equal(t, "held", res.Status)
	require.Equal(t, 1, e.provider.HoldCalls)
	holdKey := e.provider.LastHoldIdempotencyKey
	require.NotEmpty(t, holdKey)
	stored := e.res.byID[res.ID]
	require.Equal(t, holdKey, stored.ProviderRequestID)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	assertNoProviderLeak(t, rec.Body.Bytes())
	res = decodeReservation(t, rec)
	require.Equal(t, "confirmed", res.Status)
	require.Equal(t, 1, e.provider.ConfirmCalls)
	require.Equal(t, holdKey, e.provider.LastConfirmIdempotencyKey)

	// Idempotent confirm: no second provider confirm (service short-circuits)
	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, e.provider.ConfirmCalls)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	res = decodeReservation(t, rec)
	require.Equal(t, "cancelled", res.Status)
	require.Equal(t, 1, e.provider.CancelCalls)
	require.Equal(t, holdKey+":cancel", e.provider.LastCancelIdempotencyKey)

	// Idempotent cancel: already cancelled — no second provider cancel
	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "cancelled", decodeReservation(t, rec).Status)
	require.Equal(t, 1, e.provider.CancelCalls)
}

func TestBookingHTTPDuplicateIdempotencyKey(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-dup")
	body := beginBody(tt.ID, `[{"player_id":1}]`)
	key := "same-client-key"

	rec := e.do(t, http.MethodPost, "/api/v1/reservations", body, 1, key)
	require.Equal(t, http.StatusCreated, rec.Code)
	first := decodeReservation(t, rec)
	require.Equal(t, 1, e.provider.HoldCalls)
	providerKey := e.provider.LastHoldIdempotencyKey

	rec = e.do(t, http.MethodPost, "/api/v1/reservations", body, 1, key)
	require.Equal(t, http.StatusOK, rec.Code)
	second := decodeReservation(t, rec)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, 2, e.provider.HoldCalls)
	require.Equal(t, []string{providerKey, providerKey}, e.provider.HoldIdempotencyKeys)
}

func TestBookingHTTPSameIdempotencyKeyDifferentPayload(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-payload")
	key := "payload-key"

	rec := e.do(t, http.MethodPost, "/api/v1/reservations", beginBody(tt.ID, `[{"player_id":1}]`), 1, key)
	require.Equal(t, http.StatusCreated, rec.Code)
	first := decodeReservation(t, rec)
	require.Equal(t, int32(1), first.PartySize)
	providerKey := e.provider.LastHoldIdempotencyKey

	// Different party in body — still resumes the original attempt
	rec = e.do(t, http.MethodPost, "/api/v1/reservations",
		beginBody(tt.ID, `[{"player_id":1},{"guest_name":"Extra"}]`), 1, key)
	require.Equal(t, http.StatusOK, rec.Code)
	second := decodeReservation(t, rec)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, int32(1), second.PartySize)
	require.Equal(t, 2, e.provider.HoldCalls)
	require.Equal(t, providerKey, e.provider.LastHoldIdempotencyKey)
}

func TestBookingHTTPProviderReject(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-reject")
	e.provider.HoldBehavior = fakebooking.BehaviorReject

	rec := e.do(t, http.MethodPost, "/api/v1/reservations", beginBody(tt.ID, `[{"player_id":1}]`), 1, "reject-key")
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Equal(t, 1, e.provider.HoldCalls)
	require.NotEmpty(t, e.provider.LastHoldIdempotencyKey)
	require.Equal(t, entity.ReservationStatusFailed, e.res.byID[1].Status)
}

func TestBookingHTTPProviderTimeoutThenRetry(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-timeout")
	body := beginBody(tt.ID, `[{"player_id":1}]`)
	key := "timeout-key"

	e.provider.HoldBehavior = fakebooking.BehaviorTimeout
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", body, 1, key)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.Equal(t, 1, e.provider.HoldCalls)
	providerKey := e.provider.LastHoldIdempotencyKey
	require.Equal(t, entity.ReservationStatusPending, e.res.byID[1].Status)

	e.provider.HoldBehavior = fakebooking.BehaviorSuccess
	rec = e.do(t, http.MethodPost, "/api/v1/reservations", body, 1, key)
	require.Equal(t, http.StatusOK, rec.Code)
	res := decodeReservation(t, rec)
	require.Equal(t, "held", res.Status)
	require.Equal(t, int64(1), res.ID)
	require.Equal(t, 2, e.provider.HoldCalls)
	require.Equal(t, []string{providerKey, providerKey}, e.provider.HoldIdempotencyKeys)
	assertNoProviderLeak(t, rec.Body.Bytes())
}

func TestBookingHTTPLostResponseRetry(t *testing.T) {
	// Provider Hold succeeded and reservation is held, but the client never saw the 201.
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-lost")
	body := beginBody(tt.ID, `[{"player_id":1}]`)
	key := "lost-response-key"

	rec := e.do(t, http.MethodPost, "/api/v1/reservations", body, 1, key)
	require.Equal(t, http.StatusCreated, rec.Code)
	first := decodeReservation(t, rec)
	require.Equal(t, "held", first.Status)
	require.Equal(t, 1, e.provider.HoldCalls)
	providerKey := e.provider.LastHoldIdempotencyKey

	// Client retries as if the first response was lost
	rec = e.do(t, http.MethodPost, "/api/v1/reservations", body, 1, key)
	require.Equal(t, http.StatusOK, rec.Code)
	second := decodeReservation(t, rec)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, "held", second.Status)
	require.Equal(t, 2, e.provider.HoldCalls)
	require.Equal(t, []string{providerKey, providerKey}, e.provider.HoldIdempotencyKeys)
}

func TestBookingHTTPUnauthorizedCancel(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-authz-cancel")
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", beginBody(tt.ID, `[{"player_id":1}]`), 1, "authz-key")
	require.Equal(t, http.StatusCreated, rec.Code)
	res := decodeReservation(t, rec)
	require.Equal(t, 1, e.provider.HoldCalls)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 2)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Equal(t, 0, e.provider.CancelCalls)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 2)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Equal(t, 0, e.provider.ConfirmCalls)
}

func TestBookingHTTPCancelIdempotency(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-cancel-idem")
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", beginBody(tt.ID, `[{"player_id":1}]`), 1, "cancel-idem-key")
	require.Equal(t, http.StatusCreated, rec.Code)
	res := decodeReservation(t, rec)
	holdKey := e.provider.LastHoldIdempotencyKey

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "cancelled", decodeReservation(t, rec).Status)
	require.Equal(t, 1, e.provider.CancelCalls)
	require.Equal(t, holdKey+":cancel", e.provider.LastCancelIdempotencyKey)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "cancelled", decodeReservation(t, rec).Status)
	require.Equal(t, 1, e.provider.CancelCalls)
}

func TestBookingHTTPStaleCacheSearch(t *testing.T) {
	e := newBookingHTTPEnv(t)
	from := time.Date(2026, 8, 15, 13, 0, 0, 0, time.UTC)
	to := from.Add(8 * time.Hour)
	synced := from.Add(-time.Hour)
	slots := int32(4)
	_, err := e.tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: from.Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, AvailableSlots: &slots, LastSyncedAt: &synced,
	})
	require.NoError(t, err)

	e.provider.SearchErr = errors.New("provider down")
	rec := e.do(t, http.MethodGet,
		fmt.Sprintf("/api/v1/courses/1/tee-times?from=%s&to=%s", from.Format(time.RFC3339), to.Format(time.RFC3339)),
		"", 1)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, e.provider.SearchCalls)
	var list struct {
		Source   string `json:"source"`
		TeeTimes []struct {
			ID int64 `json:"id"`
		} `json:"tee_times"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Equal(t, "cache", list.Source)
	require.Len(t, list.TeeTimes, 1)
	assertNoProviderLeak(t, rec.Body.Bytes())
}

func TestBookingHTTPStaleInventoryOnHold(t *testing.T) {
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
	rec = e.do(t, http.MethodPost, "/api/v1/reservations",
		beginBody(list.TeeTimes[0].ID, `[{"player_id":1}]`), 1, "stale-hold")
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Equal(t, 1, e.provider.HoldCalls)
}

func TestBookingHTTPInvalidParty(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-party")

	t.Run("whitespace_guest_name", func(t *testing.T) {
		rec := e.do(t, http.MethodPost, "/api/v1/reservations",
			beginBody(tt.ID, `[{"guest_name":"   "}]`), 1, "party-ws")
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Equal(t, 0, e.provider.HoldCalls)
	})

	t.Run("empty_players", func(t *testing.T) {
		rec := e.do(t, http.MethodPost, "/api/v1/reservations",
			fmt.Sprintf(`{"tee_time_id":%d,"players":[]}`, tt.ID), 1, "party-empty")
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Equal(t, 0, e.provider.HoldCalls)
	})

	t.Run("party_too_large", func(t *testing.T) {
		rec := e.do(t, http.MethodPost, "/api/v1/reservations",
			beginBody(tt.ID, `[{"guest_name":"A"},{"guest_name":"B"},{"guest_name":"C"},{"guest_name":"D"},{"guest_name":"E"}]`),
			1, "party-five")
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Equal(t, 0, e.provider.HoldCalls)
	})

	t.Run("unknown_player_id", func(t *testing.T) {
		rec := e.do(t, http.MethodPost, "/api/v1/reservations",
			beginBody(tt.ID, `[{"player_id":0}]`), 1, "party-zero")
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Equal(t, 0, e.provider.HoldCalls)
	})
}

func TestBookingHTTPBodyLimits(t *testing.T) {
	e := newBookingHTTPEnv(t)

	t.Run("oversized_body_413", func(t *testing.T) {
		huge := `{"tee_time_id":1,"players":[{"guest_name":"` + strings.Repeat("x", 20*1024) + `"}]}`
		rec := e.do(t, http.MethodPost, "/api/v1/reservations", huge, 1, "huge")
		require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
		require.Equal(t, 0, e.provider.HoldCalls)
	})

	t.Run("truncated_json_400", func(t *testing.T) {
		rec := e.do(t, http.MethodPost, "/api/v1/reservations", `{"tee_time_id":1,"players":[`, 1, "trunc")
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Equal(t, 0, e.provider.HoldCalls)
	})

	t.Run("malformed_json_400", func(t *testing.T) {
		rec := e.do(t, http.MethodPost, "/api/v1/reservations", `{not-json`, 1, "bad")
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Equal(t, 0, e.provider.HoldCalls)
	})
}

func TestBookingHTTPInvalidWindow(t *testing.T) {
	e := newBookingHTTPEnv(t)
	from := time.Now().UTC()
	to := from.Add(100 * 24 * time.Hour)
	rec := e.do(t, http.MethodGet,
		fmt.Sprintf("/api/v1/courses/1/tee-times?from=%s&to=%s", from.Format(time.RFC3339), to.Format(time.RFC3339)),
		"", 1)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, 0, e.provider.SearchCalls)
}

func TestBookingHTTPMissingIdempotencyKey(t *testing.T) {
	e := newBookingHTTPEnv(t)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/reservations", strings.NewReader(`{"tee_time_id":1,"players":[{"player_id":1}]}`))
	r.Header.Set("Content-Type", "application/json")
	tok, err := auth.GenerateToken(1, 0, testJWTSecret)
	require.NoError(t, err)
	r.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, r)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, 0, e.provider.HoldCalls)
}

func TestBookingHTTPRejectsProviderFieldsInBody(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-fields")
	payload := map[string]any{
		"tee_time_id":         tt.ID,
		"provider":            "lightspeed",
		"provider_request_id": "client-should-not-set",
		"players":             []map[string]any{{"player_id": 1}},
	}
	b, _ := json.Marshal(payload)
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", string(b), 1, "fields-key")
	require.Equal(t, http.StatusCreated, rec.Code)
	assertNoProviderLeak(t, rec.Body.Bytes())
	require.NotEqual(t, "client-should-not-set", e.provider.LastHoldIdempotencyKey)
	require.Equal(t, e.res.byID[1].ProviderRequestID, e.provider.LastHoldIdempotencyKey)
}

func TestBookingHTTPCancelProviderFailure(t *testing.T) {
	e := newBookingHTTPEnv(t)
	tt := e.seedLinkedTee(t, "ls-cancel-fail")
	rec := e.do(t, http.MethodPost, "/api/v1/reservations", beginBody(tt.ID, `[{"player_id":1}]`), 1, "cancel-fail")
	require.Equal(t, http.StatusCreated, rec.Code)
	res := decodeReservation(t, rec)

	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/confirm", res.ID), "", 1)
	require.Equal(t, http.StatusOK, rec.Code)

	e.provider.CancelBehavior = fakebooking.BehaviorCancelFail
	rec = e.do(t, http.MethodPost, fmt.Sprintf("/api/v1/reservations/%d/cancel", res.ID), "", 1)
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Equal(t, 1, e.provider.CancelCalls)
	require.Equal(t, entity.ReservationStatusConfirmed, e.res.byID[res.ID].Status)
}

