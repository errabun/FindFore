package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type PlayerEventRepo struct {
	q  *sqlcgen.Queries
	db *sql.DB
}

func NewPlayerEventRepo(q *sqlcgen.Queries, db *sql.DB) *PlayerEventRepo {
	return &PlayerEventRepo{q: q, db: db}
}

func (r *PlayerEventRepo) Get(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	row, err := r.q.GetPlayerEvent(ctx, sqlcgen.GetPlayerEventParams{
		PlayerID: playerID,
		EventID:  eventID,
	})
	if err != nil {
		return nil, err
	}
	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID,
		EventID:      row.EventID,
		InviteStatus: entity.InviteStatus(row.InviteStatus),
	}, nil
}

func (r *PlayerEventRepo) Create(ctx context.Context, pe entity.PlayerEvent) (*entity.PlayerEvent, error) {
	row, err := r.q.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
		PlayerID:     pe.PlayerID,
		EventID:      pe.EventID,
		InviteStatus: int32(pe.InviteStatus),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, entity.ErrAlreadyOnEvent
		}
		return nil, err
	}
	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID,
		EventID:      row.EventID,
		InviteStatus: entity.InviteStatus(row.InviteStatus),
	}, nil
}

func (r *PlayerEventRepo) UpdateStatus(ctx context.Context, playerID, eventID int64, status entity.InviteStatus) (*entity.PlayerEvent, error) {
	row, err := r.q.UpdatePlayerEventStatus(ctx, sqlcgen.UpdatePlayerEventStatusParams{
		PlayerID:     playerID,
		EventID:      eventID,
		InviteStatus: int32(status),
	})
	if err != nil {
		return nil, err
	}
	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID,
		EventID:      row.EventID,
		InviteStatus: entity.InviteStatus(row.InviteStatus),
	}, nil
}

func (r *PlayerEventRepo) ListPlayerIDsByEventAndStatus(ctx context.Context, eventID int64, status entity.InviteStatus) ([]int64, error) {
	return r.q.ListPlayerIDsByEventAndStatus(ctx, sqlcgen.ListPlayerIDsByEventAndStatusParams{
		EventID:      eventID,
		InviteStatus: int32(status),
	})
}

func (r *PlayerEventRepo) CountAcceptedForEvent(ctx context.Context, eventID int64) (int64, error) {
	return r.q.CountAcceptedForEvent(ctx, eventID)
}

func (r *PlayerEventRepo) ClosePendingForEvent(ctx context.Context, eventID int64) error {
	return r.q.ClosePendingForEvent(ctx, eventID)
}

func (r *PlayerEventRepo) ReopenClosedForEvent(ctx context.Context, eventID int64) error {
	return r.q.ReopenClosedForEvent(ctx, eventID)
}

// JoinAccepted locks the event, enforces capacity, inserts accepted membership, and closes
// pending invites when the event becomes full — all in one transaction.
func (r *PlayerEventRepo) JoinAccepted(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin join tx: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	openSpots, err := qtx.LockEventOpenSpots(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEventMissing
		}
		return nil, fmt.Errorf("lock event %d: %w", eventID, err)
	}
	capacity := openSpots.Int32

	_, err = qtx.GetPlayerEvent(ctx, sqlcgen.GetPlayerEventParams{
		PlayerID: playerID,
		EventID:  eventID,
	})
	if err == nil {
		return nil, entity.ErrAlreadyOnEvent
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get player event: %w", err)
	}

	acceptedCount, err := qtx.CountAcceptedForEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("count accepted: %w", err)
	}
	if acceptedCount >= int64(capacity) {
		return nil, entity.ErrEventFull
	}

	row, err := qtx.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
		PlayerID:     playerID,
		EventID:      eventID,
		InviteStatus: int32(entity.InviteStatusAccepted),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, entity.ErrAlreadyOnEvent
		}
		return nil, fmt.Errorf("create player event: %w", err)
	}

	if acceptedCount+1 >= int64(capacity) {
		if err := qtx.ClosePendingForEvent(ctx, eventID); err != nil {
			return nil, fmt.Errorf("close pending: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit join tx: %w", err)
	}

	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID,
		EventID:      row.EventID,
		InviteStatus: entity.InviteStatus(row.InviteStatus),
	}, nil
}

// AcceptInvite locks the event, enforces capacity, and sets an existing membership to accepted.
// Already-accepted memberships succeed idempotently.
func (r *PlayerEventRepo) AcceptInvite(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin accept tx: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	openSpots, err := qtx.LockEventOpenSpots(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrEventMissing
		}
		return nil, fmt.Errorf("lock event %d: %w", eventID, err)
	}
	capacity := openSpots.Int32

	existing, err := qtx.GetPlayerEvent(ctx, sqlcgen.GetPlayerEventParams{
		PlayerID: playerID,
		EventID:  eventID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrPlayerEventMissing
		}
		return nil, fmt.Errorf("get player event: %w", err)
	}

	if entity.InviteStatus(existing.InviteStatus) == entity.InviteStatusAccepted {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit accept tx: %w", err)
		}
		return &entity.PlayerEvent{
			ID:           existing.ID,
			PlayerID:     existing.PlayerID,
			EventID:      existing.EventID,
			InviteStatus: entity.InviteStatusAccepted,
		}, nil
	}

	acceptedCount, err := qtx.CountAcceptedForEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("count accepted: %w", err)
	}
	if acceptedCount >= int64(capacity) {
		return nil, entity.ErrEventFull
	}

	row, err := qtx.UpdatePlayerEventStatus(ctx, sqlcgen.UpdatePlayerEventStatusParams{
		PlayerID:     playerID,
		EventID:      eventID,
		InviteStatus: int32(entity.InviteStatusAccepted),
	})
	if err != nil {
		return nil, fmt.Errorf("accept player event: %w", err)
	}

	if acceptedCount+1 >= int64(capacity) {
		if err := qtx.ClosePendingForEvent(ctx, eventID); err != nil {
			return nil, fmt.Errorf("close pending: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit accept tx: %w", err)
	}

	return &entity.PlayerEvent{
		ID:           row.ID,
		PlayerID:     row.PlayerID,
		EventID:      row.EventID,
		InviteStatus: entity.InviteStatus(row.InviteStatus),
	}, nil
}
