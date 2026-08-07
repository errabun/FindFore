package friends

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Request(
	ctx context.Context,
	actorID, playerID int32,
) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error) {
	if actorID <= 0 || playerID <= 0 {
		return nil, nil, nil, fmt.Errorf("invalid player id")
	}
	if actorID == playerID {
		return nil, nil, nil, ErrFriendshipSelf
	}

	existing, err := s.friendships.FindBetween(ctx, actorID, playerID)
	if err == nil {
		switch existing.Status {
		case entity.FriendshipStatusAccepted:
			return nil, nil, nil, ErrFriendshipAlreadyFriends
		case entity.FriendshipStatusPending:
			if existing.RequesterID == actorID {
				return nil, nil, nil, ErrFriendshipAlreadyPending
			}
			accepted, err := s.friendships.UpdateStatus(ctx, existing.ID, entity.FriendshipStatusAccepted)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("accept reverse request: %w", err)
			}
			return s.loadPair(ctx, accepted)
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil, fmt.Errorf("find friendship: %w", err)
	}

	created, err := s.friendships.Create(ctx, actorID, playerID, entity.FriendshipStatusPending)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create friend request: %w", err)
	}
	return s.loadPair(ctx, created)
}
