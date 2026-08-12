package booking

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

// ListReservationPlayers returns party members for a reservation (no authz — callers enforce ownership).
func (s *Service) ListReservationPlayers(ctx context.Context, reservationID int64) ([]entity.ReservationPlayer, error) {
	return s.reservations.ListPlayers(ctx, reservationID)
}
