package postgres

import (
	"context"
	"database/sql"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type PlayerEventRepo struct {
	q *sqlcgen.Queries
}

func NewPlayerEventRepo(q *sqlcgen.Queries) *PlayerEventRepo {
	return &PlayerEventRepo{q: q}
}

func (r *PlayerEventRepo) Get(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	row, err := r.q.GetPlayerEvent(ctx, sqlcgen.GetPlayerEventParams{
		PlayerID: sql.NullInt64{Int64: playerID, Valid: true},
		EventID:  sql.NullInt64{Int64: eventID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID.Int64,
		EventID:      row.EventID.Int64,
		InviteStatus: entity.InviteStatus(row.InviteStatus.Int32),
	}, nil
}

func (r *PlayerEventRepo) Create(ctx context.Context, pe entity.PlayerEvent) (*entity.PlayerEvent, error) {
	row, err := r.q.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
		PlayerID:     sql.NullInt64{Int64: pe.PlayerID, Valid: true},
		EventID:      sql.NullInt64{Int64: pe.EventID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: int32(pe.InviteStatus), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID.Int64,
		EventID:      row.EventID.Int64,
		InviteStatus: entity.InviteStatus(row.InviteStatus.Int32),
	}, nil
}

func (r *PlayerEventRepo) UpdateStatus(ctx context.Context, playerID, eventID int64, status entity.InviteStatus) (*entity.PlayerEvent, error) {
	row, err := r.q.UpdatePlayerEventStatus(ctx, sqlcgen.UpdatePlayerEventStatusParams{
		PlayerID:     sql.NullInt64{Int64: playerID, Valid: true},
		EventID:      sql.NullInt64{Int64: eventID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: int32(status), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID.Int64,
		EventID:      row.EventID.Int64,
		InviteStatus: entity.InviteStatus(row.InviteStatus.Int32),
	}, nil
}

func (r *PlayerEventRepo) ListPlayerIDsByEventAndStatus(ctx context.Context, eventID int64, status entity.InviteStatus) ([]int64, error) {
	rows, err := r.q.ListPlayerIDsByEventAndStatus(ctx, sqlcgen.ListPlayerIDsByEventAndStatusParams{
		EventID:      sql.NullInt64{Int64: eventID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: int32(status), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.Int64
	}
	return ids, nil
}

func (r *PlayerEventRepo) CountAcceptedForEvent(ctx context.Context, eventID int64) (int64, error) {
	return r.q.CountAcceptedForEvent(ctx, sql.NullInt64{Int64: eventID, Valid: true})
}

func (r *PlayerEventRepo) ClosePendingForEvent(ctx context.Context, eventID int64) error {
	return r.q.ClosePendingForEvent(ctx, sql.NullInt64{Int64: eventID, Valid: true})
}

func (r *PlayerEventRepo) ReopenClosedForEvent(ctx context.Context, eventID int64) error {
	return r.q.ReopenClosedForEvent(ctx, sql.NullInt64{Int64: eventID, Valid: true})
}
