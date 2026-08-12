package booking_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/ericrabun/findfore-go/internal/application/booking"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
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
	active  map[int64]int64 // teeTimeID -> reservationID
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
	existing, ok := f.byID[r.ID]
	if !ok {
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
	_ = existing
	return &cp, nil
}

func (f *fakeReservations) ListPlayers(_ context.Context, reservationID int64) ([]entity.ReservationPlayer, error) {
	return append([]entity.ReservationPlayer(nil), f.players[reservationID]...), nil
}

type fakeProvider struct {
	name                 string
	slots                []port.BookingSlot
	holdErr              error
	confirmErr           error
	cancelErr            error
	confirmedImmediately bool
	holds                int
	confirms             int
	cancels              int
}

func (f *fakeProvider) ProviderName() string { return f.name }

func (f *fakeProvider) SearchAvailability(context.Context, string, time.Time, time.Time) ([]port.BookingSlot, error) {
	return f.slots, nil
}

func (f *fakeProvider) Hold(context.Context, port.HoldRequest) (*port.HoldResult, error) {
	f.holds++
	if f.holdErr != nil {
		return nil, f.holdErr
	}
	exp := time.Now().UTC().Add(10 * time.Minute)
	return &port.HoldResult{
		ExternalReservationID: "ext-hold-1",
		HoldExpiresAt:         &exp,
		ConfirmedImmediately:  f.confirmedImmediately,
	}, nil
}

func (f *fakeProvider) Confirm(context.Context, port.ConfirmRequest) (*port.ConfirmResult, error) {
	f.confirms++
	if f.confirmErr != nil {
		return nil, f.confirmErr
	}
	return &port.ConfirmResult{ExternalReservationID: "ext-conf-1"}, nil
}

func (f *fakeProvider) Cancel(context.Context, port.CancelRequest) error {
	f.cancels++
	return f.cancelErr
}

func TestSearchAvailabilityUpsertsAndLinksProvider(t *testing.T) {
	tees := newFakeTeeTimes()
	res := newFakeReservations()
	cap := int32(4)
	slots := int32(2)
	from := time.Date(2026, 8, 15, 7, 0, 0, 0, time.UTC)
	to := from.Add(12 * time.Hour)
	provider := &fakeProvider{
		name: entity.ProviderLightspeed,
		slots: []port.BookingSlot{{
			ExternalID: "ls-1", StartsAt: from.Add(time.Hour), Holes: "18",
			Capacity: &cap, AvailableSlots: &slots, Status: entity.TeeTimeStatusAvailable,
		}},
	}
	svc := booking.NewService(tees, res, provider)

	got, err := svc.SearchAvailability(context.Background(), 9, "course-ext", from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, int64(9), got[0].CourseID)

	linked, err := tees.GetByProviderExternalID(context.Background(), entity.ProviderLightspeed, "ls-1")
	require.NoError(t, err)
	require.Equal(t, got[0].ID, linked.ID)

	// Second search updates cache, does not conflict.
	slots2 := int32(1)
	provider.slots[0].AvailableSlots = &slots2
	got2, err := svc.SearchAvailability(context.Background(), 9, "course-ext", from, to)
	require.NoError(t, err)
	require.Len(t, got2, 1)
	require.Equal(t, int32(1), *got2[0].AvailableSlots)
}

func TestBeginBookingHoldThenConfirmAndCancel(t *testing.T) {
	tees := newFakeTeeTimes()
	tt, err := tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour), Status: entity.TeeTimeStatusAvailable,
	})
	require.NoError(t, err)
	resRepo := newFakeReservations()
	provider := &fakeProvider{name: entity.ProviderLightspeed}
	svc := booking.NewService(tees, resRepo, provider)

	res, err := svc.BeginBooking(context.Background(), booking.BeginBookingInput{
		TeeTimeID: tt.ID, BookedByPlayerID: 7, PartySize: 2,
		ExternalTeeTimeID: "ls-slot", IdempotencyKey: "idem-1",
		Players: []entity.ReservationPlayer{{PlayerID: ptr(int64(7))}},
	})
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusHeld, res.Status)
	require.Equal(t, 1, provider.holds)

	updatedTee, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusHeld, updatedTee.Status)

	confirmed, err := svc.ConfirmBooking(context.Background(), res.ID, "ls-slot", "idem-2")
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, confirmed.Status)
	require.Equal(t, "ext-conf-1", confirmed.ExternalReservationID)

	cancelled, err := svc.CancelBooking(context.Background(), confirmed.ID, "idem-3")
	require.NoError(t, err)
	require.Equal(t, entity.ReservationStatusCancelled, cancelled.Status)

	freed, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusAvailable, freed.Status)
}

func TestBeginBookingProviderFailureMarksFailed(t *testing.T) {
	tees := newFakeTeeTimes()
	tt, err := tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour), Status: entity.TeeTimeStatusAvailable,
	})
	require.NoError(t, err)
	resRepo := newFakeReservations()
	provider := &fakeProvider{name: entity.ProviderLightspeed, holdErr: errors.New("slot gone")}
	svc := booking.NewService(tees, resRepo, provider)

	res, err := svc.BeginBooking(context.Background(), booking.BeginBookingInput{
		TeeTimeID: tt.ID, BookedByPlayerID: 7, PartySize: 1, ExternalTeeTimeID: "ls-slot",
	})
	require.Error(t, err)
	require.Equal(t, entity.ReservationStatusFailed, res.Status)
	require.Contains(t, res.FailureReason, "slot gone")
}

func TestCancelBookingProviderFailureKeepsConfirmed(t *testing.T) {
	tees := newFakeTeeTimes()
	tt, err := tees.Create(context.Background(), entity.TeeTime{
		CourseID: 1, StartsAt: time.Now().UTC().Add(time.Hour), Status: entity.TeeTimeStatusBooked,
	})
	require.NoError(t, err)
	resRepo := newFakeReservations()
	res, err := resRepo.Create(context.Background(), entity.Reservation{
		TeeTimeID: tt.ID, BookedByPlayerID: 1, Status: entity.ReservationStatusConfirmed,
		PartySize: 1, Provider: entity.ProviderLightspeed, ExternalReservationID: "ext-1",
	}, nil)
	require.NoError(t, err)

	provider := &fakeProvider{name: entity.ProviderLightspeed, cancelErr: errors.New("provider down")}
	svc := booking.NewService(tees, resRepo, provider)

	out, err := svc.CancelBooking(context.Background(), res.ID, "idem")
	require.Error(t, err)
	require.Equal(t, entity.ReservationStatusConfirmed, out.Status)

	still, err := tees.GetByID(context.Background(), tt.ID)
	require.NoError(t, err)
	require.Equal(t, entity.TeeTimeStatusBooked, still.Status)
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

func ptr[T any](v T) *T { return &v }
