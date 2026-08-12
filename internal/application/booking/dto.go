package booking

import (
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

var (
	ErrProviderRequired       = errors.New("booking provider is required")
	ErrTeeTimeNotFound        = errors.New("tee time not found")
	ErrCourseNotFound         = errors.New("course not found")
	ErrInvalidParty           = errors.New("invalid party size or players")
	ErrProviderLinkMissing    = errors.New("provider link missing for course or tee time")
	ErrReservationConflict    = errors.New("cannot transition reservation in its current status")
	ErrProviderOutcomeUnknown = errors.New("provider outcome unknown; retry with same request id")
	ErrProviderRejected       = errors.New("provider rejected the booking request")
)

func errf(op string, err error) error {
	return fmt.Errorf("booking.%s: %w", op, err)
}

func cancelIdempotencyKey(providerRequestID string) string {
	return providerRequestID + ":cancel"
}

func setReservationStatus(res *entity.Reservation, to string) error {
	if err := entity.TransitionReservation(res, to); err != nil {
		return err
	}
	return nil
}

func requireOwner(actorID int64, res *entity.Reservation) error {
	if res == nil {
		return entity.ErrReservationNotFound
	}
	if actorID == 0 || actorID != res.BookedByPlayerID {
		return entity.ErrReservationForbidden
	}
	return nil
}
