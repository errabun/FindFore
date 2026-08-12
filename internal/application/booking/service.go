package booking

import "github.com/ericrabun/findfore-go/internal/domain/port"

// Service coordinates FindFore tee-time cache and provider bookings.
type Service struct {
	teeTimes     port.TeeTimeRepository
	reservations port.ReservationRepository
	courses      port.CourseRepository
	provider     port.BookingProvider
}

func NewService(
	teeTimes port.TeeTimeRepository,
	reservations port.ReservationRepository,
	courses port.CourseRepository,
	provider port.BookingProvider,
) *Service {
	return &Service{
		teeTimes:     teeTimes,
		reservations: reservations,
		courses:      courses,
		provider:     provider,
	}
}
