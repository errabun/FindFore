package booking

import "github.com/ericrabun/findfore-go/internal/domain/port"

// Service coordinates FindFore tee-time cache and provider bookings.
type Service struct {
	teeTimes     port.TeeTimeRepository
	reservations port.ReservationRepository
	provider     port.BookingProvider
}

func NewService(
	teeTimes port.TeeTimeRepository,
	reservations port.ReservationRepository,
	provider port.BookingProvider,
) *Service {
	return &Service{
		teeTimes:     teeTimes,
		reservations: reservations,
		provider:     provider,
	}
}
