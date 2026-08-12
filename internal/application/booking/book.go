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
// External provider ids are resolved inside the service — never accept them from clients.
type BeginBookingInput = port.BeginBookingInput

// BeginBookingResult includes whether a new reservation row was created (HTTP 201 vs 200).
type BeginBookingResult = port.BeginBookingResult

// BeginBooking creates a pending reservation with a FindFore-owned provider_request_id,
// then asks the provider to hold. In-flight pending/held reservations for the same
// booker resume by re-invoking Hold with the stored key.
func (s *Service) BeginBooking(ctx context.Context, in BeginBookingInput) (*BeginBookingResult, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	if in.ActorID == 0 || in.TeeTimeID == 0 {
		return nil, ErrInvalidParty
	}
	partySize := int32(len(in.Players))
	if partySize <= 0 {
		return nil, ErrInvalidParty
	}
	for _, p := range in.Players {
		if p.PlayerID == nil && p.GuestName == "" {
			return nil, ErrInvalidParty
		}
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

	externalID, err := s.resolveTeeTimeExternalID(ctx, in.TeeTimeID)
	if err != nil {
		return nil, errf("BeginBooking", err)
	}

	if active, err := s.reservations.GetActiveByTeeTimeID(ctx, in.TeeTimeID); err == nil {
		if active.BookedByPlayerID == in.ActorID &&
			(active.Status == entity.ReservationStatusPending || active.Status == entity.ReservationStatusHeld) {
			res, holdErr := s.applyHold(ctx, active, externalID, in.Players)
			if holdErr != nil {
				return &BeginBookingResult{Reservation: res, Created: false}, holdErr
			}
			return &BeginBookingResult{Reservation: res, Created: false}, nil
		}
		return nil, entity.ErrActiveReservationExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, errf("BeginBooking", err)
	}

	res, err := s.reservations.Create(ctx, entity.Reservation{
		TeeTimeID:         in.TeeTimeID,
		BookedByPlayerID:  in.ActorID,
		Status:            entity.ReservationStatusPending,
		PartySize:         partySize,
		Provider:          s.provider.ProviderName(),
		ProviderRequestID: uuid.NewString(),
		QuotedPriceCents:  teeTime.PriceCents,
		QuotedCurrency:    teeTime.Currency,
	}, in.Players)
	if err != nil {
		return nil, errf("BeginBooking", err)
	}

	held, holdErr := s.applyHold(ctx, res, externalID, in.Players)
	if holdErr != nil {
		return &BeginBookingResult{Reservation: held, Created: true}, holdErr
	}
	return &BeginBookingResult{Reservation: held, Created: true}, nil
}

func (s *Service) resolveTeeTimeExternalID(ctx context.Context, teeTimeID int64) (string, error) {
	link, err := s.teeTimes.GetProviderByTeeTime(ctx, teeTimeID, s.provider.ProviderName())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrProviderLinkMissing
		}
		return "", err
	}
	return link.ExternalID, nil
}

func (s *Service) applyHold(ctx context.Context, res *entity.Reservation, externalTeeTimeID string, players []entity.ReservationPlayer) (*entity.Reservation, error) {
	if len(players) == 0 {
		var listErr error
		players, listErr = s.reservations.ListPlayers(ctx, res.ID)
		if listErr != nil {
			return nil, errf("BeginBooking", listErr)
		}
	}

	hold, err := s.provider.Hold(ctx, port.HoldRequest{
		ExternalTeeTimeID: externalTeeTimeID,
		PartySize:         res.PartySize,
		Players:           players,
		IdempotencyKey:    res.ProviderRequestID,
	})
	if err != nil {
		if errors.Is(err, ErrProviderOutcomeUnknown) {
			return res, errf("BeginBooking", err)
		}
		if setErr := setReservationStatus(res, entity.ReservationStatusFailed); setErr != nil {
			return nil, errf("BeginBooking", errors.Join(err, setErr))
		}
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
	next := entity.ReservationStatusHeld
	teeStatus := entity.TeeTimeStatusHeld
	if hold.ConfirmedImmediately {
		next = entity.ReservationStatusConfirmed
		teeStatus = entity.TeeTimeStatusBooked
	}
	if setErr := setReservationStatus(res, next); setErr != nil {
		return nil, errf("BeginBooking", setErr)
	}
	if _, err := s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, teeStatus); err != nil {
		return nil, errf("BeginBooking", err)
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
	if res.HoldExpiresAt == nil || !res.HoldExpiresAt.Before(now) {
		return res, nil
	}
	if err := setReservationStatus(res, entity.ReservationStatusFailed); err != nil {
		return nil, errf("ExpireHold", err)
	}
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
