package store

import (
	"context"
	"database/sql"
)

type PlayerWithDetails struct {
	ID       int64
	Name     string
	Phone    string
	Email    string
	Username string
	Friends  []int64
	Events   []int64
}

func GetPlayerWithDetails(ctx context.Context, q *Queries, playerID int64) (*PlayerWithDetails, error) {
	player, err := q.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	followeeIDs, err := q.ListFolloweeIDsByFollowerID(ctx, sql.NullInt32{Int32: int32(playerID), Valid: true})
	if err != nil {
		return nil, err
	}

	friends := make([]int64, len(followeeIDs))
	for i, fid := range followeeIDs {
		friends[i] = int64(fid.Int32)
	}

	eventIDs, err := q.ListAcceptedEventIDsByPlayerID(ctx, sql.NullInt64{Int64: playerID, Valid: true})
	if err != nil {
		return nil, err
	}

	events := make([]int64, len(eventIDs))
	for i, eid := range eventIDs {
		events[i] = eid.Int64
	}

	return &PlayerWithDetails{
		ID:       player.ID,
		Name:     player.Name.String,
		Phone:    player.Phone.String,
		Email:    player.Email.String,
		Username: player.Username.String,
		Friends:  friends,
		Events:   events,
	}, nil
}
