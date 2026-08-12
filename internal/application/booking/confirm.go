package booking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// ConfirmBooking finalizes a held (or pending) reservation with the provider.
func (s *Service) ConfirmBooking(ctx context.Context, reservationID int64, externalTeeTimeID, idempotencyKey string) (*entity.Reservation, error) {
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
	if res.Status == entity.ReservationStatusConfirmed {
		return res, nil
	}
	if res.Status != entity.ReservationStatusHeld && res.Status != entity.ReservationStatusPending {
		return nil, fmt.Errorf("%w: status %s", entity.ErrInvalidReservationTransition, res.Status)
	}
	if externalTeeTimeID == "" {
		return nil, ErrProviderLinkMissing
	}

	result, err := s.provider.Confirm(ctx, port.ConfirmRequest{
		ExternalTeeTimeID:     externalTeeTimeID,
		ExternalReservationID: res.ExternalReservationID,
		PartySize:             res.PartySize,
		IdempotencyKey:        idempotencyKey,
	})
	if err != nil {
		res.Status = entity.ReservationStatusFailed
		res.FailureReason = err.Error()
		updated, updateErr := s.reservations.Update(ctx, *res)
		if updateErr != nil {
			return nil, errf("ConfirmBooking", errors.Join(err, updateErr))
		}
		_, _ = s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, entity.TeeTimeStatusAvailable)
		return updated, errf("ConfirmBooking", err)
	}

	res.Status = entity.ReservationStatusConfirmed
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
