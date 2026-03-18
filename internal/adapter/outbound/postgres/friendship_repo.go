package postgres

import (
	"context"
	"database/sql"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type FriendshipRepo struct {
	q *sqlcgen.Queries
}

func NewFriendshipRepo(q *sqlcgen.Queries) *FriendshipRepo {
	return &FriendshipRepo{q: q}
}

func (r *FriendshipRepo) Find(ctx context.Context, followerID, followeeID int32) (*entity.Friendship, error) {
	row, err := r.q.FindFriendship(ctx, sqlcgen.FindFriendshipParams{
		FollowerID: sql.NullInt32{Int32: followerID, Valid: true},
		FolloweeID: sql.NullInt32{Int32: followeeID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.Friendship{
		ID:         row.ID,
		FollowerID: row.FollowerID.Int32,
		FolloweeID: row.FolloweeID.Int32,
	}, nil
}

func (r *FriendshipRepo) Create(ctx context.Context, followerID, followeeID int32) (*entity.Friendship, error) {
	row, err := r.q.CreateFriendship(ctx, sqlcgen.CreateFriendshipParams{
		FollowerID: sql.NullInt32{Int32: followerID, Valid: true},
		FolloweeID: sql.NullInt32{Int32: followeeID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.Friendship{
		ID:         row.ID,
		FollowerID: row.FollowerID.Int32,
		FolloweeID: row.FolloweeID.Int32,
	}, nil
}

func (r *FriendshipRepo) Delete(ctx context.Context, followerID, followeeID int32) error {
	return r.q.DeleteFriendship(ctx, sqlcgen.DeleteFriendshipParams{
		FollowerID: sql.NullInt32{Int32: followerID, Valid: true},
		FolloweeID: sql.NullInt32{Int32: followeeID, Valid: true},
	})
}

func (r *FriendshipRepo) ListFolloweeIDs(ctx context.Context, followerID int32) ([]int64, error) {
	rows, err := r.q.ListFolloweeIDsByFollowerID(ctx, sql.NullInt32{Int32: followerID, Valid: true})
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = int64(row.Int32)
	}
	return ids, nil
}

func (r *FriendshipRepo) ListAcceptedEventIDs(ctx context.Context, playerID int64) ([]int64, error) {
	rows, err := r.q.ListAcceptedEventIDsByPlayerID(ctx, sql.NullInt64{Int64: playerID, Valid: true})
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.Int64
	}
	return ids, nil
}
