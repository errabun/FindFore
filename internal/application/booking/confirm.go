package booking

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// ConfirmBooking finalizes a held (or pending) reservation with the provider.
// Retries reuse the reservation's provider_request_id as IdempotencyKey.
func (s *Service) ConfirmBooking(ctx context.Context, actorID, reservationID int64) (*entity.Reservation, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	res, err := s.reservations.GetByID(ctx, reservationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, errf("ConfirmBooking", err)
	}
	if err := requireOwner(actorID, res); err != nil {
		return nil, err
	}
	if res.Status == entity.ReservationStatusConfirmed {
		return res, nil
	}

	externalID, err := s.resolveTeeTimeExternalID(ctx, res.TeeTimeID)
	if err != nil {
		return nil, errf("ConfirmBooking", err)
	}

	result, err := s.provider.Confirm(ctx, port.ConfirmRequest{
		ExternalTeeTimeID:     externalID,
		ExternalReservationID: res.ExternalReservationID,
		PartySize:             res.PartySize,
		IdempotencyKey:        res.ProviderRequestID,
	})
	if err != nil {
		if errors.Is(err, ErrProviderOutcomeUnknown) {
			return res, errf("ConfirmBooking", err)
		}
		if setErr := setReservationStatus(res, entity.ReservationStatusFailed); setErr != nil {
			return nil, errf("ConfirmBooking", errors.Join(err, setErr))
		}
		res.FailureReason = err.Error()
		updated, updateErr := s.reservations.Update(ctx, *res)
		if updateErr != nil {
			return nil, errf("ConfirmBooking", errors.Join(err, updateErr))
		}
		_, _ = s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, entity.TeeTimeStatusAvailable)
		return updated, errf("ConfirmBooking", err)
	}

	if setErr := setReservationStatus(res, entity.ReservationStatusConfirmed); setErr != nil {
		return nil, errf("ConfirmBooking", setErr)
	}
	if result != nil && result.ExternalReservationID != "" {
		res.ExternalReservationID = result.ExternalReservationID
	}
	res.HoldExpiresAt = nil
	res.FailureReason = ""
	updated, err := s.reservations.Update(ctx, *res)
	if err != nil {
		return nil, errf("ConfirmBooking", err)
	}
	if _, err := s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, entity.TeeTimeStatusBooked); err != nil {
		return nil, errf("ConfirmBooking", err)
	}
	return updated, nil
}
