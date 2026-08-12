package booking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
	"github.com/google/uuid"
)

// BeginBookingInput starts a party reservation against a FindFore tee time.
type BeginBookingInput struct {
	TeeTimeID         int64
	BookedByPlayerID  int64
	PartySize         int32
	Players           []entity.ReservationPlayer
	ExternalTeeTimeID string
}

// BeginBooking creates a pending reservation with a FindFore-owned provider_request_id,
// then asks the provider to hold. In-flight pending/held reservations for the same
// booker resume by re-invoking Hold with the stored key.
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
			return s.applyHold(ctx, active, in)
		}
		return nil, entity.ErrActiveReservationExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, errf("BeginBooking", err)
	}

	res, err := s.reservations.Create(ctx, entity.Reservation{
		TeeTimeID:         in.TeeTimeID,
		BookedByPlayerID:  in.BookedByPlayerID,
		Status:            entity.ReservationStatusPending,
		PartySize:         in.PartySize,
		Provider:          s.provider.ProviderName(),
		ProviderRequestID: uuid.NewString(),
		QuotedPriceCents:  teeTime.PriceCents,
		QuotedCurrency:    teeTime.Currency,
	}, in.Players)
	if err != nil {
		return nil, errf("BeginBooking", err)
	}

	return s.applyHold(ctx, res, in)
}

func (s *Service) applyHold(ctx context.Context, res *entity.Reservation, in BeginBookingInput) (*entity.Reservation, error) {
	players := in.Players
	if len(players) == 0 {
		var listErr error
		players, listErr = s.reservations.ListPlayers(ctx, res.ID)
		if listErr != nil {
			return nil, errf("BeginBooking", listErr)
		}
	}

	hold, err := s.provider.Hold(ctx, port.HoldRequest{
		ExternalTeeTimeID: in.ExternalTeeTimeID,
		PartySize:         res.PartySize,
		Players:           players,
		IdempotencyKey:    res.ProviderRequestID,
	})
	if err != nil {
		if errors.Is(err, ErrProviderOutcomeUnknown) {
			// Leave pending/held; client retries same reservation.
			return res, errf("BeginBooking", err)
		}
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
	res.FailureReason = ""
	if hold.ConfirmedImmediately {
		res.Status = entity.ReservationStatusConfirmed
		if _, err := s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, entity.TeeTimeStatusBooked); err != nil {
			return nil, errf("BeginBooking", err)
		}
	} else {
		res.Status = entity.ReservationStatusHeld
		if _, err := s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, entity.TeeTimeStatusHeld); err != nil {
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
