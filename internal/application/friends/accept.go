package friends

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Accept(
	ctx context.Context,
	actorID int32,
	friendshipID int64,
) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error) {
	f, err := s.friendships.GetByID(ctx, friendshipID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil, ErrFriendshipNotFound
		}
		return nil, nil, nil, fmt.Errorf("get friendship: %w", err)
	}
	if f.AddresseeID != actorID {
		return nil, nil, nil, ErrFriendshipForbidden
	}
	if f.Status != entity.FriendshipStatusPending {
		return nil, nil, nil, ErrFriendshipAlreadyFriends
	}

	accepted, err := s.friendships.UpdateStatus(ctx, f.ID, entity.FriendshipStatusAccepted)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("accept friendship: %w", err)
	}
	return s.loadPair(ctx, accepted)
}
