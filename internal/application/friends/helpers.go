package friends

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) loadPair(
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
