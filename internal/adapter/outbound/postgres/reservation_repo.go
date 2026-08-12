package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type ReservationRepo struct {
	q  *sqlcgen.Queries
	db *sql.DB
}

func NewReservationRepo(q *sqlcgen.Queries, db *sql.DB) *ReservationRepo {
	return &ReservationRepo{q: q, db: db}
}

func mapReservation(row sqlcgen.Reservation) entity.Reservation {
	var hold *time.Time
	if row.HoldExpiresAt.Valid {
		t := row.HoldExpiresAt.Time
		hold = &t
	}
	return entity.Reservation{
		ID:                    row.ID,
		TeeTimeID:             row.TeeTimeID,
		BookedByPlayerID:      row.BookedByPlayerID,
		Status:                row.Status,
		PartySize:             row.PartySize,
		Provider:              row.Provider,
		ExternalReservationID: row.ExternalReservationID.String,
		HoldExpiresAt:         hold,
		FailureReason:         row.FailureReason.String,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func mapReservationPlayer(row sqlcgen.ReservationPlayer) entity.ReservationPlayer {
	var playerID *int64
	if row.PlayerID.Valid {
		v := row.PlayerID.Int64
		playerID = &v
	}
	return entity.ReservationPlayer{
		ID:            row.ID,
		ReservationID: row.ReservationID,
		PlayerID:      playerID,
		GuestName:     row.GuestName.String,
	}
}

func (r *ReservationRepo) GetByID(ctx context.Context, id int64) (*entity.Reservation, error) {
	row, err := r.q.GetReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := mapReservation(row)
	return &out, nil
}

func (r *ReservationRepo) GetActiveByTeeTimeID(ctx context.Context, teeTimeID int64) (*entity.Reservation, error) {
	row, err := r.q.GetActiveReservationByTeeTimeID(ctx, teeTimeID)
	if err != nil {
		return nil, err
	}
	out := mapReservation(row)
	return &out, nil
}

func (r *ReservationRepo) Create(ctx context.Context, res entity.Reservation, players []entity.ReservationPlayer) (*entity.Reservation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin reservation tx: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	var hold sql.NullTime
	if res.HoldExpiresAt != nil {
		hold = sql.NullTime{Time: *res.HoldExpiresAt, Valid: true}
	}
	status := res.Status
	if status == "" {
		status = entity.ReservationStatusPending
	}

	row, err := qtx.InsertReservation(ctx, sqlcgen.InsertReservationParams{
		TeeTimeID:             res.TeeTimeID,
		BookedByPlayerID:      res.BookedByPlayerID,
		Status:                status,
		PartySize:             res.PartySize,
		Provider:              res.Provider,
		ExternalReservationID: sql.NullString{String: res.ExternalReservationID, Valid: res.ExternalReservationID != ""},
		HoldExpiresAt:         hold,
		FailureReason:         sql.NullString{String: res.FailureReason, Valid: res.FailureReason != ""},
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, entity.ErrActiveReservationExists
		}
		return nil, err
	}

	for _, p := range players {
		var pid sql.NullInt64
		if p.PlayerID != nil {
			pid = sql.NullInt64{Int64: *p.PlayerID, Valid: true}
		}
		_, err := qtx.InsertReservationPlayer(ctx, sqlcgen.InsertReservationPlayerParams{
			ReservationID: row.ID,
			PlayerID:      pid,
			GuestName:     sql.NullString{String: p.GuestName, Valid: p.GuestName != ""},
		})
		if err != nil {
			return nil, fmt.Errorf("insert reservation player: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit reservation: %w", err)
	}
	out := mapReservation(row)
	return &out, nil
}

func (r *ReservationRepo) Update(ctx context.Context, res entity.Reservation) (*entity.Reservation, error) {
	var hold sql.NullTime
	if res.HoldExpiresAt != nil {
		hold = sql.NullTime{Time: *res.HoldExpiresAt, Valid: true}
	}
	row, err := r.q.UpdateReservation(ctx, sqlcgen.UpdateReservationParams{
		ID:                    res.ID,
		Status:                res.Status,
		ExternalReservationID: sql.NullString{String: res.ExternalReservationID, Valid: res.ExternalReservationID != ""},
		HoldExpiresAt:         hold,
		FailureReason:         sql.NullString{String: res.FailureReason, Valid: res.FailureReason != ""},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, err
	}
	out := mapReservation(row)
	return &out, nil
}

func (r *ReservationRepo) ListPlayers(ctx context.Context, reservationID int64) ([]entity.ReservationPlayer, error) {
	rows, err := r.q.ListReservationPlayers(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	out := make([]entity.ReservationPlayer, len(rows))
	for i, row := range rows {
		out[i] = mapReservationPlayer(row)
	}
	return out, nil
}
