package port

import (
	"context"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

// BookingSlot is a provider tee-sheet slot before/after FindFore upsert.
type BookingSlot struct {
	ExternalID     string
	StartsAt       time.Time
	Holes          string
	Capacity       *int32
	AvailableSlots *int32
	PriceCents     *int32
	Currency       string
	Status         string
}

// HoldRequest asks a provider to hold inventory for a party.
type HoldRequest struct {
	ExternalTeeTimeID string
	PartySize         int32
	Players           []entity.ReservationPlayer
	IdempotencyKey    string
}

// HoldResult is a successful (or soft) hold from the provider.
type HoldResult struct {
	ExternalReservationID string
	HoldExpiresAt         *time.Time
	// ConfirmedImmediately is true when the provider has no separate hold step.
	ConfirmedImmediately bool
}

// ConfirmRequest finalizes a held booking (or books directly when no hold).
type ConfirmRequest struct {
	ExternalTeeTimeID     string
	ExternalReservationID string
	PartySize             int32
	IdempotencyKey        string
}

// ConfirmResult is a successful confirmation.
type ConfirmResult struct {
	ExternalReservationID string
}

// CancelRequest asks the provider to release a confirmed (or held) booking.
type CancelRequest struct {
	ExternalReservationID string
	IdempotencyKey        string
}

// BookingProvider is the outbound port for tee-sheet search and booking.
// Adapters (Lightspeed, ForeUP, …) map vendor DTOs; domain never imports them.
type BookingProvider interface {
	ProviderName() string
	SearchAvailability(ctx context.Context, courseExternalID string, from, to time.Time) ([]BookingSlot, error)
	Hold(ctx context.Context, req HoldRequest) (*HoldResult, error)
	Confirm(ctx context.Context, req ConfirmRequest) (*ConfirmResult, error)
	Cancel(ctx context.Context, req CancelRequest) error
}
