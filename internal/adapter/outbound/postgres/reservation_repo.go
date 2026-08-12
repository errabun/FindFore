package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/google/uuid"
)

type ReservationRepo struct {
	q  *sqlcgen.Queries
	db *sql.DB
}

func NewReservationRepo(q *sqlcgen.Queries, db *sql.DB) *ReservationRepo {
	return &ReservationRepo{q: q, db: db}
}

func mapReservationRow(
	id, teeTimeID, bookedBy int64,
	status string,
	partySize int32,
	provider string,
	externalID sql.NullString,
	holdExpires sql.NullTime,
	failure sql.NullString,
	providerRequestID uuid.UUID,
	quotedPrice sql.NullInt32,
	quotedCurrency sql.NullString,
	clientIdempotency sql.NullString,
	createdAt, updatedAt time.Time,
) entity.Reservation {
	var hold *time.Time
	if holdExpires.Valid {
		t := holdExpires.Time
		hold = &t
	}
	return entity.Reservation{
		ID:                    id,
		TeeTimeID:             teeTimeID,
		BookedByPlayerID:      bookedBy,
		Status:                status,
		PartySize:             partySize,
		Provider:              provider,
		ExternalReservationID: externalID.String,
		HoldExpiresAt:         hold,
		FailureReason:         failure.String,
		ProviderRequestID:     providerRequestID.String(),
		QuotedPriceCents:      int32Ptr(quotedPrice),
		QuotedCurrency:        quotedCurrency.String,
		ClientIdempotencyKey:  clientIdempotency.String,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
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
	out := mapReservationRow(
		row.ID, row.TeeTimeID, row.BookedByPlayerID, row.Status, row.PartySize, row.Provider,
		row.ExternalReservationID, row.HoldExpiresAt, row.FailureReason, row.ProviderRequestID,
		row.QuotedPriceCents, row.QuotedCurrency, row.ClientIdempotencyKey, row.CreatedAt, row.UpdatedAt,
	)
	return &out, nil
}

func (r *ReservationRepo) GetActiveByTeeTimeID(ctx context.Context, teeTimeID int64) (*entity.Reservation, error) {
	row, err := r.q.GetActiveReservationByTeeTimeID(ctx, teeTimeID)
	if err != nil {
		return nil, err
	}
	out := mapReservationRow(
		row.ID, row.TeeTimeID, row.BookedByPlayerID, row.Status, row.PartySize, row.Provider,
		row.ExternalReservationID, row.HoldExpiresAt, row.FailureReason, row.ProviderRequestID,
		row.QuotedPriceCents, row.QuotedCurrency, row.ClientIdempotencyKey, row.CreatedAt, row.UpdatedAt,
	)
	return &out, nil
}

func (r *ReservationRepo) GetByClientIdempotency(ctx context.Context, bookedByPlayerID int64, clientIdempotencyKey string) (*entity.Reservation, error) {
	row, err := r.q.GetReservationByClientIdempotency(ctx, sqlcgen.GetReservationByClientIdempotencyParams{
		BookedByPlayerID:     bookedByPlayerID,
		ClientIdempotencyKey: sql.NullString{String: clientIdempotencyKey, Valid: clientIdempotencyKey != ""},
	})
	if err != nil {
		return nil, err
	}
	out := mapReservationRow(
		row.ID, row.TeeTimeID, row.BookedByPlayerID, row.Status, row.PartySize, row.Provider,
		row.ExternalReservationID, row.HoldExpiresAt, row.FailureReason, row.ProviderRequestID,
		row.QuotedPriceCents, row.QuotedCurrency, row.ClientIdempotencyKey, row.CreatedAt, row.UpdatedAt,
	)
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
	reqID, err := uuid.Parse(res.ProviderRequestID)
	if err != nil {
		return nil, fmt.Errorf("provider_request_id: %w", err)
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
		ProviderRequestID:     reqID,
		QuotedPriceCents:      nullInt32Ptr(res.QuotedPriceCents),
		QuotedCurrency:        sql.NullString{String: res.QuotedCurrency, Valid: res.QuotedCurrency != ""},
		ClientIdempotencyKey:  sql.NullString{String: res.ClientIdempotencyKey, Valid: res.ClientIdempotencyKey != ""},
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
	out := mapReservationRow(
		row.ID, row.TeeTimeID, row.BookedByPlayerID, row.Status, row.PartySize, row.Provider,
		row.ExternalReservationID, row.HoldExpiresAt, row.FailureReason, row.ProviderRequestID,
		row.QuotedPriceCents, row.QuotedCurrency, row.ClientIdempotencyKey, row.CreatedAt, row.UpdatedAt,
	)
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
		QuotedPriceCents:      nullInt32Ptr(res.QuotedPriceCents),
		QuotedCurrency:        sql.NullString{String: res.QuotedCurrency, Valid: res.QuotedCurrency != ""},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, err
	}
	out := mapReservationRow(
		row.ID, row.TeeTimeID, row.BookedByPlayerID, row.Status, row.PartySize, row.Provider,
		row.ExternalReservationID, row.HoldExpiresAt, row.FailureReason, row.ProviderRequestID,
		row.QuotedPriceCents, row.QuotedCurrency, row.ClientIdempotencyKey, row.CreatedAt, row.UpdatedAt,
	)
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
