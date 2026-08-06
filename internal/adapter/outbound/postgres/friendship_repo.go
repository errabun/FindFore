package postgres

import (
	"context"
	"database/sql"
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

func mapFriendshipRow(id int64, requesterID, addresseeID sql.NullInt32, status int32) *entity.Friendship {
	return &entity.Friendship{
		ID:          id,
		RequesterID: requesterID.Int32,
		AddresseeID: addresseeID.Int32,
		Status:      entity.FriendshipStatus(status),
	}
}

func (r *FriendshipRepo) GetByID(ctx context.Context, id int64) (*entity.Friendship, error) {
	row, err := r.q.GetFriendshipByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapFriendshipRow(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) Find(ctx context.Context, requesterID, addresseeID int32) (*entity.Friendship, error) {
	row, err := r.q.FindFriendship(ctx, sqlcgen.FindFriendshipParams{
		RequesterID: sql.NullInt32{Int32: requesterID, Valid: true},
		AddresseeID: sql.NullInt32{Int32: addresseeID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapFriendshipRow(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) FindBetween(ctx context.Context, playerA, playerB int32) (*entity.Friendship, error) {
	row, err := r.q.FindFriendshipBetween(ctx, sqlcgen.FindFriendshipBetweenParams{
		RequesterID: sql.NullInt32{Int32: playerA, Valid: true},
		AddresseeID: sql.NullInt32{Int32: playerB, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapFriendshipRow(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) Create(ctx context.Context, requesterID, addresseeID int32, status entity.FriendshipStatus) (*entity.Friendship, error) {
	row, err := r.q.CreateFriendship(ctx, sqlcgen.CreateFriendshipParams{
		RequesterID: sql.NullInt32{Int32: requesterID, Valid: true},
		AddresseeID: sql.NullInt32{Int32: addresseeID, Valid: true},
		Status:      int32(status),
	})
	if err != nil {
		return nil, err
	}
	return mapFriendshipRow(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) UpdateStatus(ctx context.Context, id int64, status entity.FriendshipStatus) (*entity.Friendship, error) {
	row, err := r.q.UpdateFriendshipStatus(ctx, sqlcgen.UpdateFriendshipStatusParams{
		ID:     id,
		Status: int32(status),
	})
	if err != nil {
		return nil, err
	}
	return mapFriendshipRow(row.ID, row.RequesterID, row.AddresseeID, row.Status), nil
}

func (r *FriendshipRepo) DeleteByID(ctx context.Context, id int64) error {
	return r.q.DeleteFriendshipByID(ctx, id)
}

func (r *FriendshipRepo) ListAcceptedFriendIDs(ctx context.Context, playerID int32) ([]int64, error) {
	rows, err := r.q.ListAcceptedFriendIDs(ctx, sql.NullInt32{Int32: playerID, Valid: true})
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, id := range rows {
		ids[i] = int64(id)
	}
	return ids, nil
}

func (r *FriendshipRepo) ListIncomingPending(ctx context.Context, addresseeID int32) ([]entity.Friendship, error) {
	rows, err := r.q.ListIncomingPendingFriendships(ctx, sql.NullInt32{Int32: addresseeID, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]entity.Friendship, len(rows))
	for i, row := range rows {
		out[i] = *mapFriendshipRow(row.ID, row.RequesterID, row.AddresseeID, row.Status)
	}
	return out, nil
}

func (r *FriendshipRepo) ListOutgoingPending(ctx context.Context, requesterID int32) ([]entity.Friendship, error) {
	rows, err := r.q.ListOutgoingPendingFriendships(ctx, sql.NullInt32{Int32: requesterID, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]entity.Friendship, len(rows))
	for i, row := range rows {
		out[i] = *mapFriendshipRow(row.ID, row.RequesterID, row.AddresseeID, row.Status)
	}
	return out, nil
}

func (r *FriendshipRepo) ListAcceptedEventIDs(ctx context.Context, playerID int64) ([]int64, error) {
	rows, err := r.q.ListAcceptedEventIDsByPlayerID(ctx, sql.NullInt64{Int64: playerID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list accepted event ids: %w", err)
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.Int64
	}
	return ids, nil
}
