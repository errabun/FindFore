package booking

import (
	"errors"
	"fmt"
)

var (
	ErrProviderRequired       = errors.New("booking provider is required")
	ErrTeeTimeNotFound        = errors.New("tee time not found")
	ErrInvalidParty           = errors.New("invalid party size or players")
	ErrProviderLinkMissing    = errors.New("tee time has no provider external id for this provider")
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
