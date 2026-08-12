package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type EventRepo struct {
	q  *sqlcgen.Queries
	db *sql.DB
}

func NewEventRepo(q *sqlcgen.Queries, db *sql.DB) *EventRepo {
	return &EventRepo{q: q, db: db}
}

func nullTeeTimeID(id *int64) sql.NullInt64 {
	if id == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *id, Valid: true}
}

func teeTimeIDPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

func (r *EventRepo) GetByID(ctx context.Context, id int64) (*entity.Event, error) {
	row, err := r.q.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.Event{
		ID:              row.ID,
		CourseID:        int32(row.CourseID),
		PlannedStartsAt: row.PlannedStartsAt,
		TeeTimeID:       teeTimeIDPtr(row.TeeTimeID),
		OpenSpots:       row.OpenSpots.Int32,
		NumberOfHoles:   row.NumberOfHoles.String,
		Private:         row.Private.Bool,
		HostID:          int32(row.HostID),
	}, nil
}

func (r *EventRepo) GetDetailsByID(ctx context.Context, id int64) (*entity.EventWithDetails, error) {
	row, err := r.q.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.EventWithDetails{
		ID:              row.ID,
		CourseName:      row.CourseName.String,
		CourseTimezone:  row.CourseTimezone.String,
		PlannedStartsAt: row.PlannedStartsAt,
		TeeTimeID:       teeTimeIDPtr(row.TeeTimeID),
		OpenSpots:       row.OpenSpots.Int32,
		NumberOfHoles:   row.NumberOfHoles.String,
		Private:         row.Private.Bool,
		HostName:        row.HostName.String,
		HostID:          int32(row.HostID),
	}, nil
}

func (r *EventRepo) ListAllIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.q.ListAllEvents(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return ids, nil
}

func (r *EventRepo) ListPublicIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.q.ListPublicEvents(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return ids, nil
}

func (r *EventRepo) ListIDsByPlayerID(ctx context.Context, playerID int64) ([]int64, error) {
	rows, err := r.q.ListEventsByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return ids, nil
}

func (r *EventRepo) ListFriendsAvailableIDs(ctx context.Context, followerID int32, playerID int64) ([]int64, error) {
	return r.q.ListFriendsAvailableEventIDs(ctx, sqlcgen.ListFriendsAvailableEventIDsParams{
		RequesterID: int64(followerID),
		PlayerID:    playerID,
	})
}

func createEventParams(e entity.Event) sqlcgen.CreateEventParams {
	return sqlcgen.CreateEventParams{
		CourseID:        int64(e.CourseID),
		OpenSpots:       sql.NullInt32{Int32: e.OpenSpots, Valid: true},
		NumberOfHoles:   sql.NullString{String: e.NumberOfHoles, Valid: true},
		Private:         sql.NullBool{Bool: e.Private, Valid: true},
		HostID:          int64(e.HostID),
		PlannedStartsAt: e.PlannedStartsAt,
		TeeTimeID:       nullTeeTimeID(e.TeeTimeID),
	}
}

func (r *EventRepo) Create(ctx context.Context, e entity.Event) (int64, error) {
	if e.PlannedStartsAt.IsZero() {
		return 0, fmt.Errorf("planned_starts_at is required")
	}
	row, err := r.q.CreateEvent(ctx, createEventParams(e))
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *EventRepo) CreateWithInvites(ctx context.Context, e entity.Event, invitees []int64) (int64, error) {
	if e.PlannedStartsAt.IsZero() {
		return 0, fmt.Errorf("planned_starts_at is required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	event, err := qtx.CreateEvent(ctx, createEventParams(e))
	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}

	_, err = qtx.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
		PlayerID:     int64(e.HostID),
		EventID:      event.ID,
		InviteStatus: int32(entity.InviteStatusAccepted),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create host player_event: %w", err)
	}

	if e.Private {
		for _, inviteeID := range invitees {
			if inviteeID == int64(e.HostID) {
				continue
			}
			_, err = qtx.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
				PlayerID:     inviteeID,
				EventID:      event.ID,
				InviteStatus: int32(entity.InviteStatusPending),
			})
			if err != nil {
				return 0, fmt.Errorf("failed to create invitee player_event: %w", err)
			}
		}
	} else {
		playerIDs, err := qtx.ListPlayersExceptHost(ctx, int64(e.HostID))
		if err != nil {
			return 0, fmt.Errorf("failed to list players: %w", err)
		}
		for _, pid := range playerIDs {
			_, err = qtx.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
				PlayerID:     pid,
				EventID:      event.ID,
				InviteStatus: int32(entity.InviteStatusPending),
			})
			if err != nil {
				return 0, fmt.Errorf("failed to create player_event: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return event.ID, nil
}

func (r *EventRepo) Update(ctx context.Context, e entity.Event) error {
	if e.PlannedStartsAt.IsZero() {
		return fmt.Errorf("planned_starts_at is required")
	}
	return r.q.UpdateEvent(ctx, sqlcgen.UpdateEventParams{
		ID:              e.ID,
		CourseID:        int64(e.CourseID),
		OpenSpots:       sql.NullInt32{Int32: e.OpenSpots, Valid: true},
		NumberOfHoles:   sql.NullString{String: e.NumberOfHoles, Valid: true},
		Private:         sql.NullBool{Bool: e.Private, Valid: true},
		PlannedStartsAt: e.PlannedStartsAt,
		TeeTimeID:       nullTeeTimeID(e.TeeTimeID),
	})
}

func (r *EventRepo) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteEvent(ctx, id)
}

func (r *EventRepo) DeletePast(ctx context.Context) error {
	return r.q.DeletePastEvents(ctx)
}
