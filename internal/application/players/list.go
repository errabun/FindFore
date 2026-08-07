package players

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) List(ctx context.Context) ([]entity.PlayerWithDetails, error) {
	players, err := s.players.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}

	result := make([]entity.PlayerWithDetails, 0, len(players))
	for _, p := range players {
		details, err := s.GetWithDetails(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("get player details for %d: %w", p.ID, err)
		}
		result = append(result, *details)
	}
	return result, nil
}

func (s *Service) GetWithDetails(ctx context.Context, id int64) (*entity.PlayerWithDetails, error) {
	player, err := s.players.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get player %d: %w", id, err)
	}

	friendIDs, err := s.friendships.ListAcceptedFriendIDs(ctx, int32(player.ID))
	if err != nil {
		return nil, fmt.Errorf("list friend IDs for %d: %w", id, err)
	}

	eventIDs, err := s.friendships.ListAcceptedEventIDs(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("list accepted event IDs for %d: %w", id, err)
	}

	return &entity.PlayerWithDetails{
		ID:       player.ID,
		Name:     player.Name,
		Phone:    player.Phone,
		Email:    player.Email,
		Username: player.Username,
		Friends:  friendIDs,
		Events:   eventIDs,
	}, nil
}
