package booking_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/fakebooking"
	"github.com/ericrabun/findfore-go/internal/application/booking"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeTeeTimes struct {
	byID       map[int64]*entity.TeeTime
	byProvider map[string]*entity.TeeTime
	providers  map[string]*entity.TeeTimeProvider
	byTeeProv  map[string]*entity.TeeTimeProvider
	nextID     int64
}

func newFakeTeeTimes() *fakeTeeTimes {
	return &fakeTeeTimes{
		byID:       map[int64]*entity.TeeTime{},
		byProvider: map[string]*entity.TeeTime{},
		providers:  map[string]*entity.TeeTimeProvider{},
		byTeeProv:  map[string]*entity.TeeTimeProvider{},
		nextID:     1,
	}
}

func providerKey(provider, externalID string) string { return provider + ":" + externalID }
func teeKey(teeTimeID int64, provider string) string {
	return fmt.Sprintf("%s/%d", provider, teeTimeID)
}

func (f *fakeTeeTimes) GetByID(_ context.Context, id int64) (*entity.TeeTime, error) {
	t, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *t
	return &cp, nil
}

func (f *fakeTeeTimes) ListByCourseAndWindow(_ context.Context, courseID int64, from, to time.Time) ([]entity.TeeTime, error) {
	var out []entity.TeeTime
	for _, t := range f.byID {
		if t.CourseID == courseID && !t.StartsAt.Before(from) && t.StartsAt.Before(to) {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (f *fakeTeeTimes) GetByProviderExternalID(_ context.Context, provider, externalID string) (*entity.TeeTime, error) {
	t, ok := f.byProvider[providerKey(provider, externalID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *t
	return &cp, nil
}

func (f *fakeTeeTimes) Create(_ context.Context, t entity.TeeTime) (*entity.TeeTime, error) {
	t.ID = f.nextID
	f.nextID++
	cp := t
	f.byID[t.ID] = &cp
	return &t, nil
}

func (f *fakeTeeTimes) UpdateCache(_ context.Context, t entity.TeeTime) (*entity.TeeTime, error) {
	if _, ok := f.byID[t.ID]; !ok {
		return nil, sql.ErrNoRows
	}
	cp := t
	f.byID[t.ID] = &cp
	for k, v := range f.byProvider {
		if v.ID == t.ID {
			f.byProvider[k] = &cp
		}
	}
	return &t, nil
}

func (f *fakeTeeTimes) UpdateStatus(_ context.Context, id int64, status string) (*entity.TeeTime, error) {
	t, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	t.Status = status
	cp := *t
	return &cp, nil
}

func (f *fakeTeeTimes) GetProvider(_ context.Context, provider, externalID string) (*entity.TeeTimeProvider, error) {
	p, ok := f.providers[providerKey(provider, externalID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (f *fakeTeeTimes) GetProviderByTeeTime(_ context.Context, teeTimeID int64, provider string) (*entity.TeeTimeProvider, error) {
	p, ok := f.byTeeProv[teeKey(teeTimeID, provider)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (f *fakeTeeTimes) LinkProvider(_ context.Context, teeTimeID int64, provider, externalID string) error {
	key := providerKey(provider, externalID)
	if existing, ok := f.providers[key]; ok {
		if existing.TeeTimeID == teeTimeID {
			return nil
		}
		return entity.ErrProviderTeeTimeConflict
	}
	p := &entity.TeeTimeProvider{
		ID: int64(len(f.providers) + 1), TeeTimeID: teeTimeID, Provider: provider, ExternalID: externalID,
	}
	f.providers[key] = p
	f.byTeeProv[teeKey(teeTimeID, provider)] = p
	if t, ok := f.byID[teeTimeID]; ok {
		f.byProvider[key] = t
	}
	return nil
}

type fakeReservations struct {
	byID    map[int64]*entity.Reservation
	active  map[int64]int64
	nextID  int64
	players map[int64][]entity.ReservationPlayer
}

func newFakeReservations() *fakeReservations {
	return &fakeReservations{
		byID: map[int64]*entity.Reservation{}, active: map[int64]int64{},
		players: map[int64][]entity.ReservationPlayer{}, nextID: 1,
	}
}

func (f *fakeReservations) GetByID(_ context.Context, id int64) (*entity.Reservation, error) {
	r, ok := f.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *r
	return &cp, nil
}

func (f *fakeReservations) GetActiveByTeeTimeID(_ context.Context, teeTimeID int64) (*entity.Reservation, error) {
	id, ok := f.active[teeTimeID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return f.GetByID(context.Background(), id)
}

func (f *fakeReservations) GetByClientIdempotency(_ context.Context, bookedByPlayerID int64, clientIdempotencyKey string) (*entity.Reservation, error) {
	for _, r := range f.byID {
		if r.BookedByPlayerID == bookedByPlayerID && r.ClientIdempotencyKey == clientIdempotencyKey {
			cp := *r
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (f *fakeReservations) Create(_ context.Context, r entity.Reservation, players []entity.ReservationPlayer) (*entity.Reservation, error) {
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

func (f *fakeReservations) Update(_ context.Context, r entity.Reservation) (*entity.Reservation, error) {
	if _, ok := f.byID[r.ID]; !ok {
		return nil, entity.ErrReservationNotFound
	}
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

func (f *fakeReservations) ListPlayers(_ context.Context, reservationID int64) ([]entity.ReservationPlayer, error) {
	return append([]entity.ReservationPlayer(nil), f.players[reservationID]...), nil
}

type fakeCourses struct {
	providers map[string]*entity.CourseProvider
}

func newFakeCourses() *fakeCourses {
	return &fakeCourses{providers: map[string]*entity.CourseProvider{}}
}

func (f *fakeCourses) List(context.Context) ([]entity.Course, error) { return nil, nil }
func (f *fakeCourses) GetByID(context.Context, int64) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeCourses) GetByNameAndCity(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeCourses) GetByProviderExternalID(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeCourses) Create(context.Context, entity.Course) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeCourses) GetProvider(context.Context, string, string) (*entity.CourseProvider, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeCourses) GetProviderByCourse(_ context.Context, courseID int64, provider string) (*entity.CourseProvider, error) {
	p, ok := f.providers[teeKey(courseID, provider)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}
func (f *fakeCourses) LinkProvider(_ context.Context, courseID int64, provider, externalID string) error {
	f.providers[teeKey(courseID, provider)] = &entity.CourseProvider{
		CourseID: courseID, Provider: provider, ExternalID: externalID,
	}
	return nil
}

func seedLinkedTee(t *testing.T, tees *fakeTeeTimes, providerName, externalID string, price int32) *entity.TeeTime {
	t.Helper()
	tt, err := tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour),
		Status: entity.TeeTimeStatusAvailable, PriceCents: &price, Currency: "USD",
	})
	require.NoError(t, err)
	require.NoError(t, tees.LinkProvider(context.Background(), tt.ID, providerName, externalID))
	return tt
}

type fakePlayers struct{}

func (fakePlayers) List(context.Context) ([]entity.Player, error) { return nil, nil }
func (fakePlayers) GetByID(_ context.Context, id int64) (*entity.Player, error) {
	if id <= 0 {
		return nil, sql.ErrNoRows
	}
	return &entity.Player{ID: id, Name: "Player"}, nil
}
func (fakePlayers) GetByEmail(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (fakePlayers) GetByUsername(context.Context, string) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (fakePlayers) Create(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (fakePlayers) Update(context.Context, entity.Player) (*entity.Player, error) {
	return nil, sql.ErrNoRows
}
func (fakePlayers) GetPasswordByID(context.Context, int64) (string, error) { return "", sql.ErrNoRows }
func (fakePlayers) UpdatePassword(context.Context, int64, string) error    { return sql.ErrNoRows }
func (fakePlayers) GetTokenVersion(context.Context, int64) (int32, error)  { return 0, nil }
func (fakePlayers) ListIDsExcept(context.Context, int64) ([]int64, error)  { return nil, nil }

func beginIn(actor, teeID int64) booking.BeginBookingInput {
	pid := actor
	return booking.BeginBookingInput{
		ActorID: actor, TeeTimeID: teeID,
		ClientIdempotencyKey: uuid.NewString(),
		Players:              []entity.ReservationPlayer{{PlayerID: &pid, GuestName: "Eric"}},
	}
}

func newSvc(tees *fakeTeeTimes, res *fakeReservations, courses *fakeCourses, provider port.BookingProvider) *booking.Service {
	return booking.NewService(tees, res, courses, fakePlayers{}, provider)
}

func TestSearchAvailabilityResolvesCourseProvider(t *testing.T) {
	tees := newFakeTeeTimes()
	res := newFakeReservations()
	courses := newFakeCourses()
	require.NoError(t, courses.LinkProvider(context.Background(), 9, entity.ProviderLightspeed, "course-ext"))
	cap, slots, price := int32(4), int32(2), int32(6500)
	from := time.Date(2026, 8, 15, 7, 0, 0, 0, time.UTC)
	to := from.Add(12 * time.Hour)
	provider := fakebooking.New(entity.ProviderLightspeed)
	provider.Slots = []port.BookingSlot{{
		ExternalID: "ls-1", StartsAt: from.Add(time.Hour), Holes: "18",
		Capacity: &cap, AvailableSlots: &slots, PriceCents: &price, Currency: "USD",
		Status: entity.TeeTimeStatusAvailable,
	}}
	svc := newSvc(tees, res, courses, provider)

	got, err := svc.SearchAvailability(context.Background(), 9, from, to, 0)
	require.NoError(t, err)
	require.Equal(t, port.AvailabilitySourceProvider, got.Source)
	require.Len(t, got.TeeTimes, 1)
	require.NotNil(t, got.TeeTimes[0].LastSyncedAt)
}

func TestHappyHoldConfirmCancel(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedLinkedTee(t, tees, entity.ProviderLightspeed, "ls-slot", 6500)
	resRepo := newFakeReservations()
	provider := fakebooking.New(entity.ProviderLightspeed)
	svc := newSvc(tees, resRepo, newFakeCourses(), provider)

	out, err := svc.BeginBooking(context.Background(), beginIn(7, tt.ID))
	require.NoError(t, err)
	require.True(t, out.Created)
	require.Equal(t, entity.ReservationStatusHeld, out.Reservation.Status)

	confirmed, err := svc.ConfirmBooking(context.Background(), 7, out.Reservation.ID)
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, confirmed.Status)

	cancelled, err := svc.CancelBooking(context.Background(), 7, confirmed.ID)
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusCancelled, cancelled.Status)
}

func TestConfirmForbiddenForOtherActor(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedLinkedTee(t, tees, entity.ProviderLightspeed, "ls-slot", 6500)
	svc := newSvc(tees, newFakeReservations(), newFakeCourses(), fakebooking.New(entity.ProviderLightspeed))

	out, err := svc.BeginBooking(context.Background(), beginIn(7, tt.ID))
	require.NoError(t, err)
	_, err = svc.ConfirmBooking(context.Background(), 99, out.Reservation.ID)
	require.ErrorIs(t, err, entity.ErrReservationForbidden)
}

func TestHoldTimeoutLeavesPendingAndRetryReusesKey(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedLinkedTee(t, tees, entity.ProviderLightspeed, "ls-slot", 6500)
	provider := fakebooking.New(entity.ProviderLightspeed)
	provider.HoldBehavior = fakebooking.BehaviorTimeout
	svc := newSvc(tees, newFakeReservations(), newFakeCourses(), provider)

	out, err := svc.BeginBooking(context.Background(), beginIn(7, tt.ID))
	require.ErrorIs(t, err, booking.ErrProviderOutcomeUnknown)
	require.Equal(t, entity.ReservationStatusPending, out.Reservation.Status)
	reqID := out.Reservation.ProviderRequestID

	provider.HoldBehavior = fakebooking.BehaviorSuccess
	out2, err := svc.BeginBooking(context.Background(), beginIn(7, tt.ID))
	require.NoError(t, err)
	require.False(t, out2.Created)
	require.Equal(t, reqID, out2.Reservation.ProviderRequestID)
}

func TestHoldRejectMarksFailed(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedLinkedTee(t, tees, entity.ProviderLightspeed, "ls-slot", 6500)
	provider := fakebooking.New(entity.ProviderLightspeed)
	provider.HoldBehavior = fakebooking.BehaviorReject
	svc := newSvc(tees, newFakeReservations(), newFakeCourses(), provider)

	out, err := svc.BeginBooking(context.Background(), beginIn(7, tt.ID))
	require.ErrorIs(t, err, booking.ErrProviderRejected)
	require.Equal(t, entity.ReservationStatusFailed, out.Reservation.Status)
}

func TestConfirmOnlyProvider(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedLinkedTee(t, tees, entity.ProviderForeUP, "fu-slot", 7000)
	provider := fakebooking.New(entity.ProviderForeUP)
	provider.ConfirmOnly = true
	svc := newSvc(tees, newFakeReservations(), newFakeCourses(), provider)

	out, err := svc.BeginBooking(context.Background(), beginIn(7, tt.ID))
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, out.Reservation.Status)
}

func TestCancelProviderFailureKeepsConfirmed(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedLinkedTee(t, tees, entity.ProviderLightspeed, "ls-slot", 6500)
	provider := fakebooking.New(entity.ProviderLightspeed)
	svc := newSvc(tees, newFakeReservations(), newFakeCourses(), provider)

	out, err := svc.BeginBooking(context.Background(), beginIn(7, tt.ID))
	require.NoError(t, err)
	confirmed, err := svc.ConfirmBooking(context.Background(), 7, out.Reservation.ID)
	require.NoError(t, err)

	provider.CancelBehavior = fakebooking.BehaviorCancelFail
	got, err := svc.CancelBooking(context.Background(), 7, confirmed.ID)
	require.Error(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, got.Status)
}

func TestIllegalTransitionBlocked(t *testing.T) {
	res := &entity.Reservation{Status: entity.ReservationStatusFailed}
	err := entity.TransitionReservation(res, entity.ReservationStatusConfirmed)
	require.ErrorIs(t, err, entity.ErrInvalidReservationTransition)
}
