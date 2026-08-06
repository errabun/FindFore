package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

var (
	ErrFriendshipNotFound      = errors.New("friendship not found")
	ErrFriendshipForbidden     = errors.New("not allowed to modify this friendship")
	ErrFriendshipSelf          = errors.New("cannot create friendship with yourself")
	ErrFriendshipAlreadyFriends = errors.New("already friends")
	ErrFriendshipAlreadyPending = errors.New("friend request already pending")
)

type FriendshipService struct {
	friendships port.FriendshipRepository
	playerSvc   port.PlayerService
}

func NewFriendshipService(friendships port.FriendshipRepository, playerSvc port.PlayerService) *FriendshipService {
	return &FriendshipService{friendships: friendships, playerSvc: playerSvc}
}

func (s *FriendshipService) loadPair(
	ctx context.Context,
	f *entity.Friendship,
) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error) {
	requester, err := s.playerSvc.GetWithDetails(ctx, int64(f.RequesterID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get requester details: %w", err)
	}
	addressee, err := s.playerSvc.GetWithDetails(ctx, int64(f.AddresseeID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get addressee details: %w", err)
	}
	return f, requester, addressee, nil
}

func (s *FriendshipService) Request(
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
			// Reverse pending: they already requested us — accept that request.
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

func (s *FriendshipService) Accept(
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

func (s *FriendshipService) Decline(ctx context.Context, actorID int32, friendshipID int64) error {
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

func (s *FriendshipService) CancelOrUnfriend(ctx context.Context, actorID int32, friendshipID int64) error {
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

func (s *FriendshipService) ListIncomingRequests(ctx context.Context, actorID int32) ([]entity.Friendship, error) {
	return s.friendships.ListIncomingPending(ctx, actorID)
}

func (s *FriendshipService) ListOutgoingPendingIDs(ctx context.Context, actorID int32) ([]int64, error) {
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

func (s *FriendshipService) ListAccepted(ctx context.Context, actorID int32) ([]entity.Friendship, error) {
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
