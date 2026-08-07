package friends

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Decline(ctx context.Context, actorID int32, friendshipID int64) error {
	f, err := s.friendships.GetByID(ctx, friendshipID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrFriendshipNotFound
		}
		return fmt.Errorf("get friendship: %w", err)
	}
	if f.AddresseeID != actorID {
		return ErrFriendshipForbidden
	}
	if f.Status != entity.FriendshipStatusPending {
		return ErrFriendshipForbidden
	}
	if err := s.friendships.DeleteByID(ctx, f.ID); err != nil {
		return fmt.Errorf("decline friendship: %w", err)
	}
	return nil
}
