package booking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// CancelBooking asks the provider to cancel a confirmed/held reservation.
// On provider failure the reservation stays confirmed/held (inventory not freed).
func (s *Service) CancelBooking(ctx context.Context, reservationID int64, idempotencyKey string) (*entity.Reservation, error) {
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
	if res.Status == entity.ReservationStatusCancelled {
		return res, nil
	}
	if res.Status != entity.ReservationStatusConfirmed &&
		res.Status != entity.ReservationStatusHeld &&
		res.Status != entity.ReservationStatusPending {
		return nil, fmt.Errorf("%w: status %s", entity.ErrInvalidReservationTransition, res.Status)
	}

	if res.ExternalReservationID != "" {
		if err := s.provider.Cancel(ctx, port.CancelRequest{
			ExternalReservationID: res.ExternalReservationID,
			IdempotencyKey:        idempotencyKey,
		}); err != nil {
			// Flow 11: stay confirmed/held; do not free inventory.
			return res, errf("CancelBooking", err)
		}
	}

	res.Status = entity.ReservationStatusCancelled
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
