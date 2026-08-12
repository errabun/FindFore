package entity

import (
	"errors"
	"time"
)

// TeeTime statuses mirror chk_tee_times_status.
const (
	TeeTimeStatusAvailable = "available"
	TeeTimeStatusHeld      = "held"
	TeeTimeStatusBooked    = "booked"
	TeeTimeStatusCancelled = "cancelled"
)

// TeeTime is FindFore's normalized tee-sheet slot (provider inventory is cached here).
type TeeTime struct {
	ID             int64
	CourseID       int64
	StartsAt       time.Time
	Holes          string
	Status         string
	Capacity       *int32
	AvailableSlots *int32
	PriceCents     *int32
	Currency       string
}

// TeeTimeProvider maps a vendor slot identity onto a canonical tee time.
type TeeTimeProvider struct {
	ID         int64
	TeeTimeID  int64
	Provider   string
	ExternalID string
}

// ErrProviderTeeTimeConflict is returned when (provider, external_id) is already
// linked to a different tee time.
var ErrProviderTeeTimeConflict = errors.New("provider external id already linked to another tee time")
