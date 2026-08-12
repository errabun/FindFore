package entity

import (
	"errors"
	"fmt"
	"time"
)

// Reservation statuses mirror chk_reservations_status.
const (
	ReservationStatusPending   = "pending"
	ReservationStatusHeld      = "held"
	ReservationStatusConfirmed = "confirmed"
	ReservationStatusCancelled = "cancelled"
	ReservationStatusFailed    = "failed"
)

// Reservation is a party booking against a tee time (not a social event).
type Reservation struct {
	ID                    int64
	TeeTimeID             int64
	BookedByPlayerID      int64
	Status                string
	PartySize             int32
	Provider              string
	ExternalReservationID string
	HoldExpiresAt         *time.Time
	FailureReason         string
	ProviderRequestID     string // FindFore-generated UUID; persisted before provider calls
	QuotedPriceCents      *int32
	QuotedCurrency        string
	ClientIdempotencyKey  string // Client Idempotency-Key; UNIQUE(booked_by, key) when set
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// ReservationPlayer is a party member (FindFore player and/or guest name).
// Both may be set: player_id = FindFore identity, guest_name = name submitted to provider.
type ReservationPlayer struct {
	ID            int64
	ReservationID int64
	PlayerID      *int64
	GuestName     string
}

// ErrActiveReservationExists is returned when a non-terminal reservation already
// exists for the tee time.
var ErrActiveReservationExists = errors.New("active reservation already exists for tee time")

// ErrReservationNotFound is returned when a reservation id does not exist.
var ErrReservationNotFound = errors.New("reservation not found")

// ErrInvalidReservationTransition is returned for illegal status changes.
var ErrInvalidReservationTransition = errors.New("invalid reservation status transition")

// ErrReservationForbidden is returned when the actor is not allowed to mutate the reservation.
var ErrReservationForbidden = errors.New("reservation action forbidden for this player")

// IsTerminalReservation reports whether status is terminal (retry needs a new row).
func IsTerminalReservation(status string) bool {
	return status == ReservationStatusCancelled || status == ReservationStatusFailed
}

// IsActiveReservation reports whether status blocks another booking on the same tee time.
func IsActiveReservation(status string) bool {
	switch status {
	case ReservationStatusPending, ReservationStatusHeld, ReservationStatusConfirmed:
		return true
	default:
		return false
	}
}

// CanTransitionReservation reports whether from→to is a legal reservation status edge.
// Same-status is not a transition (callers handle idempotent no-ops before calling).
func CanTransitionReservation(from, to string) bool {
	if from == to {
		return false
	}
	switch from {
	case ReservationStatusPending:
		return to == ReservationStatusHeld || to == ReservationStatusConfirmed || to == ReservationStatusFailed
	case ReservationStatusHeld:
		return to == ReservationStatusConfirmed || to == ReservationStatusFailed || to == ReservationStatusCancelled
	case ReservationStatusConfirmed:
		return to == ReservationStatusCancelled
	default:
		return false
	}
}

// TransitionReservation sets res.Status to to when the edge is legal.
func TransitionReservation(res *Reservation, to string) error {
	if res == nil {
		return ErrReservationNotFound
	}
	if !CanTransitionReservation(res.Status, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidReservationTransition, res.Status, to)
	}
	res.Status = to
	return nil
}
