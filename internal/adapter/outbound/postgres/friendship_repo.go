package postgres

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type FriendshipRepo struct {
	q *sqlcgen.Queries
}

func NewFriendshipRepo(q *sqlcgen.Queries) *FriendshipRepo {
	return &FriendshipRepo{q: q}
}

func mapFriendship(id, requesterID, addresseeID int64, status int32) *entity.Friendship {
	return &entity.Friendship{
		ID:          id,
		RequesterID: int32(requesterID),
		AddresseeID: int32(addresseeID),
		Status:      entity.FriendshipStatus(status),
	}
}

func (r *FriendshipRepo) GetByID(ctx context.Context, id int64) (*entity.Friendship, error) {
	row, err := r.q.GetFriendshipByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapFriendship(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) Find(ctx context.Context, requesterID, addresseeID int32) (*entity.Friendship, error) {
	row, err := r.q.FindFriendship(ctx, sqlcgen.FindFriendshipParams{
		RequesterID: int64(requesterID),
		AddresseeID: int64(addresseeID),
	})
	if err != nil {
		return nil, err
	}
	return mapFriendship(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) FindBetween(ctx context.Context, playerA, playerB int32) (*entity.Friendship, error) {
	row, err := r.q.FindFriendshipBetween(ctx, sqlcgen.FindFriendshipBetweenParams{
		RequesterID: int64(playerA),
		AddresseeID: int64(playerB),
	})
	if err != nil {
		return nil, err
	}
	return mapFriendship(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) Create(ctx context.Context, requesterID, addresseeID int32, status entity.FriendshipStatus) (*entity.Friendship, error) {
	row, err := r.q.CreateFriendship(ctx, sqlcgen.CreateFriendshipParams{
		RequesterID: int64(requesterID),
		AddresseeID: int64(addresseeID),
		Status:      int32(status),
	})
	if err != nil {
		return nil, err
	}
	return mapFriendship(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) UpdateStatus(ctx context.Context, id int64, status entity.FriendshipStatus) (*entity.Friendship, error) {
	row, err := r.q.UpdateFriendshipStatus(ctx, sqlcgen.UpdateFriendshipStatusParams{
		ID:     id,
		Status: int32(status),
	})
	if err != nil {
		return nil, err
	}
	return mapFriendship(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) DeleteByID(ctx context.Context, id int64) error {
	return r.q.DeleteFriendshipByID(ctx, id)
}

func (r *FriendshipRepo) ListAcceptedFriendIDs(ctx context.Context, playerID int32) ([]int64, error) {
	return r.q.ListAcceptedFriendIDs(ctx, int64(playerID))
}

func (r *FriendshipRepo) ListIncomingPending(ctx context.Context, addresseeID int32) ([]entity.Friendship, error) {
	rows, err := r.q.ListIncomingPendingFriendships(ctx, int64(addresseeID))
	if err != nil {
		return nil, err
	}
	out := make([]entity.Friendship, len(rows))
	for i, row := range rows {
		out[i] = *mapFriendship(row.ID, row.RequesterID, row.AddresseeID, row.Status)
	}
	return out, nil
}

func (r *FriendshipRepo) ListOutgoingPending(ctx context.Context, requesterID int32) ([]entity.Friendship, error) {
	rows, err := r.q.ListOutgoingPendingFriendships(ctx, int64(requesterID))
	if err != nil {
		return nil, err
	}
	out := make([]entity.Friendship, len(rows))
	for i, row := range rows {
		out[i] = *mapFriendship(row.ID, row.RequesterID, row.AddresseeID, row.Status)
	}
	return out, nil
}

func (r *FriendshipRepo) ListAcceptedEventIDs(ctx context.Context, playerID int64) ([]int64, error) {
	rows, err := r.q.ListAcceptedEventIDsByPlayerID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("list accepted event ids: %w", err)
	}
	return rows, nil
}
