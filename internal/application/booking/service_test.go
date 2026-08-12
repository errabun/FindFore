package booking_test

import (
	"context"
	"database/sql"
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
	nextID     int64
}

func newFakeTeeTimes() *fakeTeeTimes {
	return &fakeTeeTimes{
		byID:       map[int64]*entity.TeeTime{},
		byProvider: map[string]*entity.TeeTime{},
		providers:  map[string]*entity.TeeTimeProvider{},
		nextID:     1,
	}
}

func providerKey(provider, externalID string) string {
	return provider + ":" + externalID
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

func (f *fakeTeeTimes) LinkProvider(_ context.Context, teeTimeID int64, provider, externalID string) error {
	key := providerKey(provider, externalID)
	if existing, ok := f.providers[key]; ok {
		if existing.TeeTimeID == teeTimeID {
			return nil
		}
		return entity.ErrProviderTeeTimeConflict
	}
	f.providers[key] = &entity.TeeTimeProvider{
		ID: int64(len(f.providers) + 1), TeeTimeID: teeTimeID, Provider: provider, ExternalID: externalID,
	}
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
		byID:    map[int64]*entity.Reservation{},
		active:  map[int64]int64{},
		players: map[int64][]entity.ReservationPlayer{},
		nextID:  1,
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

func (f *fakeReservations) Create(_ context.Context, r entity.Reservation, players []entity.ReservationPlayer) (*entity.Reservation, error) {
	if _, ok := f.active[r.TeeTimeID]; ok && entity.IsActiveReservation(r.Status) {
		return nil, entity.ErrActiveReservationExists
	}
	if r.ProviderRequestID == "" {
		r.ProviderRequestID = uuid.NewString()
	}
	r.ID = f.nextID
	f.nextID++
	now := time.Now().UTC()
	r.CreatedAt = now
	r.UpdatedAt = now
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

func seedTee(t *testing.T, tees *fakeTeeTimes, price int32) *entity.TeeTime {
	t.Helper()
	tt, err := tees.Create(context.Background(), entity.TeeTime{
		CourseID:   1,
		StartsAt:   time.Now().UTC().Add(time.Hour),
		Status:     entity.TeeTimeStatusAvailable,
		PriceCents: &price,
		Currency:   "USD",
	})
	require.NoError(t, err)
	return tt
}

func beginInput(ttID int64) booking.BeginBookingInput {
	pid := int64(7)
	return booking.BeginBookingInput{
		TeeTimeID:         ttID,
		BookedByPlayerID:  7,
		PartySize:         2,
		ExternalTeeTimeID: "ls-slot",
		Players:           []entity.ReservationPlayer{{PlayerID: &pid, GuestName: "Eric"}},
	}
}

func TestSearchAvailabilitySetsLastSyncedAtAndQuotedReady(t *testing.T) {
	tees := newFakeTeeTimes()
	res := newFakeReservations()
	cap := int32(4)
	slots := int32(2)
	price := int32(6500)
	from := time.Date(2026, 8, 15, 7, 0, 0, 0, time.UTC)
	to := from.Add(12 * time.Hour)
	provider := fakebooking.New(entity.ProviderLightspeed)
	provider.Slots = []port.BookingSlot{{
		ExternalID: "ls-1", StartsAt: from.Add(time.Hour), Holes: "18",
		Capacity: &cap, AvailableSlots: &slots, PriceCents: &price, Currency: "USD",
		Status: entity.TeeTimeStatusAvailable,
	}}
	svc := booking.NewService(tees, res, provider)

	got, err := svc.SearchAvailability(context.Background(), 9, "course-ext", from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].LastSyncedAt)
	require.Equal(t, int32(6500), *got[0].PriceCents)
}

func TestHappyHoldConfirmCancel(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedTee(t, tees, 6500)
	resRepo := newFakeReservations()
	provider := fakebooking.New(entity.ProviderLightspeed)
	svc := booking.NewService(tees, resRepo, provider)

	res, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusHeld, res.Status)
	require.NotEmpty(t, res.ProviderRequestID)
	require.NotNil(t, res.QuotedPriceCents)
	require.Equal(t, int32(6500), *res.QuotedPriceCents)
	require.Equal(t, "USD", res.QuotedCurrency)
	require.Equal(t, 1, provider.HoldCalls)

	updatedTee, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusHeld, updatedTee.Status)

	confirmed, err := svc.ConfirmBooking(context.Background(), res.ID, "ls-slot")
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, confirmed.Status)

	cancelled, err := svc.CancelBooking(context.Background(), confirmed.ID)
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusCancelled, cancelled.Status)

	freed, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusAvailable, freed.Status)
}

func TestHoldProviderRejectMarksFailed(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedTee(t, tees, 6500)
	resRepo := newFakeReservations()
	provider := fakebooking.New(entity.ProviderLightspeed)
	provider.HoldBehavior = fakebooking.BehaviorReject
	svc := booking.NewService(tees, resRepo, provider)

	res, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.Error(t, err)
	require.ErrorIs(t, err, booking.ErrProviderRejected)
	require.Equal(t, entity.ReservationStatusFailed, res.Status)

	still, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusAvailable, still.Status)

	// New attempt after failure is allowed (new provider_request_id).
	provider.HoldBehavior = fakebooking.BehaviorSuccess
	res2, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusHeld, res2.Status)
	require.NotEqual(t, res.ProviderRequestID, res2.ProviderRequestID)
}

func TestHoldTimeoutLeavesPendingAndRetryReusesKey(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedTee(t, tees, 6500)
	resRepo := newFakeReservations()
	provider := fakebooking.New(entity.ProviderLightspeed)
	provider.HoldBehavior = fakebooking.BehaviorTimeout
	svc := booking.NewService(tees, resRepo, provider)

	res, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.Error(t, err)
	require.ErrorIs(t, err, booking.ErrProviderOutcomeUnknown)
	require.Equal(t, entity.ReservationStatusPending, res.Status)
	require.Equal(t, 1, provider.HoldCalls)
	reqID := res.ProviderRequestID

	provider.HoldBehavior = fakebooking.BehaviorSuccess
	res2, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusHeld, res2.Status)
	require.Equal(t, reqID, res2.ProviderRequestID)
	require.Equal(t, res.ID, res2.ID)
	require.Equal(t, 2, provider.HoldCalls)

	// Idempotent third call returns cached hold without inventing a new external id.
	res3, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)
	require.Equal(t, res2.ExternalReservationID, res3.ExternalReservationID)
	require.Equal(t, 3, provider.HoldCalls)
}

func TestDuplicateBeginWhilePendingResumesSameRow(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedTee(t, tees, 6500)
	resRepo := newFakeReservations()
	provider := fakebooking.New(entity.ProviderLightspeed)
	svc := booking.NewService(tees, resRepo, provider)

	first, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)

	second, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, first.ProviderRequestID, second.ProviderRequestID)
	require.Equal(t, 2, provider.HoldCalls)
}

func TestConfirmOnlyProviderSkipsHeld(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedTee(t, tees, 7000)
	resRepo := newFakeReservations()
	provider := fakebooking.New(entity.ProviderForeUP)
	provider.ConfirmOnly = true
	svc := booking.NewService(tees, resRepo, provider)

	res, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, res.Status)

	booked, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusBooked, booked.Status)
}

func TestCancelProviderFailureKeepsConfirmed(t *testing.T) {
	tees := newFakeTeeTimes()
	tt := seedTee(t, tees, 6500)
	resRepo := newFakeReservations()
	provider := fakebooking.New(entity.ProviderLightspeed)
	svc := booking.NewService(tees, resRepo, provider)

	res, err := svc.BeginBooking(context.Background(), beginInput(tt.ID))
	require.NoError(t, err)
	confirmed, err := svc.ConfirmBooking(context.Background(), res.ID, "ls-slot")
	require.NoError(t, err)

	provider.CancelBehavior = fakebooking.BehaviorCancelFail
	out, err := svc.CancelBooking(context.Background(), confirmed.ID)
	require.Error(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, out.Status)

	still, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusBooked, still.Status)
}

func TestStaleInventoryRejectAfterSearch(t *testing.T) {
	tees := newFakeTeeTimes()
	resRepo := newFakeReservations()
	price := int32(6500)
	from := time.Date(2026, 8, 15, 7, 0, 0, 0, time.UTC)
	to := from.Add(12 * time.Hour)
	provider := fakebooking.New(entity.ProviderLightspeed)
	provider.Slots = []port.BookingSlot{{
		ExternalID: "ls-stale", StartsAt: from.Add(time.Hour),
		PriceCents: &price, Currency: "USD", Status: entity.TeeTimeStatusAvailable,
	}}
	svc := booking.NewService(tees, resRepo, provider)

	got, err := svc.SearchAvailability(context.Background(), 1, "course-ext", from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].LastSyncedAt)

	provider.GoneExternalIDs["ls-stale"] = true
	res, err := svc.BeginBooking(context.Background(), booking.BeginBookingInput{
		TeeTimeID: got[0].ID, BookedByPlayerID: 7, PartySize: 1, ExternalTeeTimeID: "ls-stale",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, booking.ErrProviderRejected)
	require.Equal(t, entity.ReservationStatusFailed, res.Status)
}

func TestLinkProviderConflict(t *testing.T) {
	tees := newFakeTeeTimes()
	a, err := tees.Create(context.Background(), entity.TeeTime{CourseID: 1, StartsAt: time.Now().UTC(), Status: entity.TeeTimeStatusAvailable})
	require.NoError(t, err)
	b, err := tees.Create(context.Background(), entity.TeeTime{CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour), Status: entity.TeeTimeStatusAvailable})
	require.NoError(t, err)
	require.NoError(t, tees.LinkProvider(context.Background(), a.ID, entity.ProviderLightspeed, "same"))
	err = tees.LinkProvider(context.Background(), b.ID, entity.ProviderLightspeed, "same")
	require.ErrorIs(t, err, entity.ErrProviderTeeTimeConflict)
}
