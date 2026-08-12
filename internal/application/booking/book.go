package booking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
// then asks the provider to hold. Client Idempotency-Key resumes an existing attempt;
// in-flight pending/held for the same booker also resume Hold with the stored key.
func (s *Service) BeginBooking(ctx context.Context, in BeginBookingInput) (*BeginBookingResult, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	if in.ActorID == 0 || in.TeeTimeID == 0 {
		return nil, ErrInvalidParty
	}
	key := strings.TrimSpace(in.ClientIdempotencyKey)
	if key == "" || len(key) > maxClientIdempotencyLen {
		return nil, ErrInvalidParty
	}
	in.ClientIdempotencyKey = key

	players, err := s.normalizeAndValidateParty(ctx, in.Players)
	if err != nil {
		return nil, err
	}
	in.Players = players
	partySize := int32(len(players))

	if existing, err := s.reservations.GetByClientIdempotency(ctx, in.ActorID, key); err == nil {
		if existing.Status == entity.ReservationStatusPending || existing.Status == entity.ReservationStatusHeld {
			externalID, linkErr := s.resolveTeeTimeExternalID(ctx, existing.TeeTimeID)
			if linkErr != nil {
				return nil, errf("BeginBooking", linkErr)
			}
			res, holdErr := s.applyHold(ctx, existing, externalID, nil)
			if holdErr != nil {
				return &BeginBookingResult{Reservation: res, Created: false}, holdErr
			}
			return &BeginBookingResult{Reservation: res, Created: false}, nil
		}
		return &BeginBookingResult{Reservation: existing, Created: false}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, errf("BeginBooking", err)
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
	if teeTime.AvailableSlots != nil && partySize > *teeTime.AvailableSlots {
		return nil, fmt.Errorf("%w: party exceeds available slots", ErrReservationConflict)
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
		TeeTimeID:            in.TeeTimeID,
		BookedByPlayerID:     in.ActorID,
		Status:               entity.ReservationStatusPending,
		PartySize:            partySize,
		Provider:             s.provider.ProviderName(),
		ProviderRequestID:    uuid.NewString(),
		QuotedPriceCents:     teeTime.PriceCents,
		QuotedCurrency:       teeTime.Currency,
		ClientIdempotencyKey: key,
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

func (s *Service) normalizeAndValidateParty(ctx context.Context, in []entity.ReservationPlayer) ([]entity.ReservationPlayer, error) {
	n := len(in)
	if n < 1 || n > maxPartySize {
		return nil, ErrInvalidParty
	}
	out := make([]entity.ReservationPlayer, 0, n)
	for _, p := range in {
		name := strings.TrimSpace(p.GuestName)
		if p.PlayerID == nil && name == "" {
			return nil, ErrInvalidParty
		}
		if p.PlayerID != nil {
			if s.players == nil {
				return nil, ErrInvalidParty
			}
			if _, err := s.players.GetByID(ctx, *p.PlayerID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, ErrInvalidParty
				}
				return nil, errf("BeginBooking", err)
			}
		}
		out = append(out, entity.ReservationPlayer{PlayerID: p.PlayerID, GuestName: name})
	}
	return out, nil
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
	if res.Status != next {
		if setErr := setReservationStatus(res, next); setErr != nil {
			return nil, errf("BeginBooking", setErr)
		}
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
