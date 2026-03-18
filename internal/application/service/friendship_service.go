package service

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type FriendshipService struct {
	friendships port.FriendshipRepository
	playerSvc   port.PlayerService
}

func NewFriendshipService(friendships port.FriendshipRepository, playerSvc port.PlayerService) *FriendshipService {
	return &FriendshipService{friendships: friendships, playerSvc: playerSvc}
}

func (s *FriendshipService) FindOrCreate(ctx context.Context, followerID, followeeID int32) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error) {
	var friendship *entity.Friendship

	existing, err := s.friendships.Find(ctx, followerID, followeeID)
	if err == nil {
		friendship = existing
	} else {
		created, err := s.friendships.Create(ctx, followerID, followeeID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create friendship: %w", err)
		}
		friendship = created
	}

	followerDetails, err := s.playerSvc.GetWithDetails(ctx, int64(friendship.FollowerID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get follower details: %w", err)
	}

	followeeDetails, err := s.playerSvc.GetWithDetails(ctx, int64(friendship.FolloweeID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get followee details: %w", err)
	}

	return friendship, followerDetails, followeeDetails, nil
}

func (s *FriendshipService) Delete(ctx context.Context, followerID, followeeID int32) error {
	return s.friendships.Delete(ctx, followerID, followeeID)
}
