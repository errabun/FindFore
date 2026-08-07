package friends

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) CancelOrUnfriend(ctx context.Context, actorID int32, friendshipID int64) error {
	f, err := s.friendships.GetByID(ctx, friendshipID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrFriendshipNotFound
		}
		return fmt.Errorf("get friendship: %w", err)
	}

	switch f.Status {
	case entity.FriendshipStatusPending:
		if f.RequesterID != actorID {
			return ErrFriendshipForbidden
		}
	case entity.FriendshipStatusAccepted:
		if f.RequesterID != actorID && f.AddresseeID != actorID {
			return ErrFriendshipForbidden
		}
	default:
		return ErrFriendshipForbidden
	}

	if err := s.friendships.DeleteByID(ctx, f.ID); err != nil {
		return fmt.Errorf("delete friendship: %w", err)
	}
	return nil
}
