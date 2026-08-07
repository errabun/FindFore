package friends

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) ListIncomingRequests(ctx context.Context, actorID int32) ([]entity.Friendship, error) {
	return s.friendships.ListIncomingPending(ctx, actorID)
}

func (s *Service) ListOutgoingPendingIDs(ctx context.Context, actorID int32) ([]int64, error) {
	rows, err := s.friendships.ListOutgoingPending(ctx, actorID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = int64(row.AddresseeID)
	}
	return ids, nil
}

func (s *Service) ListAccepted(ctx context.Context, actorID int32) ([]entity.Friendship, error) {
	ids, err := s.friendships.ListAcceptedFriendIDs(ctx, actorID)
	if err != nil {
		return nil, err
	}
	out := make([]entity.Friendship, 0, len(ids))
	for _, friendID := range ids {
		f, err := s.friendships.FindBetween(ctx, actorID, int32(friendID))
		if err != nil {
			continue
		}
		if f.Status == entity.FriendshipStatusAccepted {
			out = append(out, *f)
		}
	}
	return out, nil
}
