package booking

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// CancelBooking asks the provider to cancel a confirmed/held/pending reservation.
// On provider failure the reservation stays in its prior status (inventory not freed).
func (s *Service) CancelBooking(ctx context.Context, actorID, reservationID int64) (*entity.Reservation, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	res, err := s.reservations.GetByID(ctx, reservationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, errf("CancelBooking", err)
	}
	if err := requireOwner(actorID, res); err != nil {
		return nil, err
	}
	if res.Status == entity.ReservationStatusCancelled {
		return res, nil
	}

	if res.ExternalReservationID != "" || res.ProviderRequestID != "" {
		if err := s.provider.Cancel(ctx, port.CancelRequest{
			ExternalReservationID: res.ExternalReservationID,
			IdempotencyKey:        cancelIdempotencyKey(res.ProviderRequestID),
		}); err != nil {
			// Flow 11 / unknown: stay confirmed/held; do not free inventory.
			return res, errf("CancelBooking", err)
		}
	}

	if setErr := setReservationStatus(res, entity.ReservationStatusCancelled); setErr != nil {
		return nil, errf("CancelBooking", setErr)
	}
	res.FailureReason = ""
	updated, err := s.reservations.Update(ctx, *res)
	if err != nil {
		return nil, errf("CancelBooking", err)
	}
	if _, err := s.teeTimes.UpdateStatus(ctx, res.TeeTimeID, entity.TeeTimeStatusAvailable); err != nil {
		return nil, errf("CancelBooking", err)
	}
	return updated, nil
}
