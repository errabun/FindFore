package entity

import (
	"errors"
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
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// ReservationPlayer is a party member (FindFore player and/or guest name).
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
