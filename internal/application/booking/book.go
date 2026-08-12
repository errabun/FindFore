package booking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// BeginBookingInput starts a party reservation against a FindFore tee time.
type BeginBookingInput struct {
	TeeTimeID         int64
	BookedByPlayerID  int64
	PartySize         int32
	Players           []entity.ReservationPlayer
	ExternalTeeTimeID string // required until reverse lookup is exposed on reads
	IdempotencyKey    string
}

// BeginBooking creates a pending/held reservation and asks the provider to hold
// inventory when supported. If the provider confirms immediately (no hold API),
// the reservation moves to confirmed.
func (s *Service) BeginBooking(ctx context.Context, in BeginBookingInput) (*entity.Reservation, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	if in.PartySize <= 0 || in.TeeTimeID == 0 || in.BookedByPlayerID == 0 {
		return nil, ErrInvalidParty
	}
	if in.ExternalTeeTimeID == "" {
		return nil, ErrProviderLinkMissing
	}

	teeTime, err := s.teeTimes.GetByID(ctx, in.TeeTimeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeeTimeNotFound
		}
		return nil, errf("BeginBooking", err)
	}
	if teeTime.Status == entity.TeeTimeStatusCancelled || teeTime.Status == entity.TeeTimeStatusBooked {
		return nil, fmt.Errorf("%w: tee time status %s", ErrReservationConflict, teeTime.Status)
	}

	if active, err := s.reservations.GetActiveByTeeTimeID(ctx, in.TeeTimeID); err == nil {
		if active.BookedByPlayerID == in.BookedByPlayerID &&
			(active.Status == entity.ReservationStatusPending || active.Status == entity.ReservationStatusHeld) {
			return active, nil
		}
		return nil, entity.ErrActiveReservationExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, errf("BeginBooking", err)
	}

	res, err := s.reservations.Create(ctx, entity.Reservation{
		TeeTimeID:        in.TeeTimeID,
		BookedByPlayerID: in.BookedByPlayerID,
		Status:           entity.ReservationStatusPending,
		PartySize:        in.PartySize,
		Provider:         s.provider.ProviderName(),
	}, in.Players)
	if err != nil {
		return nil, errf("BeginBooking", err)
	}

	hold, err := s.provider.Hold(ctx, port.HoldRequest{
		ExternalTeeTimeID: in.ExternalTeeTimeID,
		PartySize:         in.PartySize,
		Players:           in.Players,
		IdempotencyKey:    in.IdempotencyKey,
	})
	if err != nil {
		res.Status = entity.ReservationStatusFailed
		res.FailureReason = err.Error()
		updated, updateErr := s.reservations.Update(ctx, *res)
		if updateErr != nil {
			return nil, errf("BeginBooking", errors.Join(err, updateErr))
		}
		return updated, errf("BeginBooking", err)
	}

	res.ExternalReservationID = hold.ExternalReservationID
	res.HoldExpiresAt = hold.HoldExpiresAt
	if hold.ConfirmedImmediately {
		res.Status = entity.ReservationStatusConfirmed
		if _, err := s.teeTimes.UpdateStatus(ctx, in.TeeTimeID, entity.TeeTimeStatusBooked); err != nil {
			return nil, errf("BeginBooking", err)
		}
	} else {
		res.Status = entity.ReservationStatusHeld
		if _, err := s.teeTimes.UpdateStatus(ctx, in.TeeTimeID, entity.TeeTimeStatusHeld); err != nil {
			return nil, errf("BeginBooking", err)
		}
	}
	updated, err := s.reservations.Update(ctx, *res)
	if err != nil {
		return nil, errf("BeginBooking", err)
	}
	return updated, nil
}

// ExpireHold marks an expired hold as failed and frees the tee time cache status.
func (s *Service) ExpireHold(ctx context.Context, reservationID int64, now time.Time) (*entity.Reservation, error) {
	res, err := s.reservations.GetByID(ctx, reservationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, errf("ExpireHold", err)
	}
	if res.Status != entity.ReservationStatusHeld {
		return nil, fmt.Errorf("%w: status %s", entity.ErrInvalidReservationTransition, res.Status)
	}
	if res.HoldExpiresAt == nil || !res.HoldExpiresAt.Before(now) {
		return res, nil
	}
	res.Status = entity.ReservationStatusFailed
	res.FailureReason = "hold expired"
	updated, err := s.reservations.Update(ctx, *res)
	if err != nil {
		return nil, errf("ExpireHold", err)
	}
	if _, err := s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, entity.TeeTimeStatusAvailable); err != nil {
		return nil, errf("ExpireHold", err)
	}
	return updated, nil
}
