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

func (r *EventRepo) GetByID(ctx context.Context, id int64) (*entity.Event, error) {
	row, err := r.q.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.Event{
		ID:            row.ID,
		CourseID:      row.CourseID.Int32,
		Date:          row.Date.String,
		TeeTime:       row.TeeTime.String,
		OpenSpots:     row.OpenSpots.Int32,
		NumberOfHoles: row.NumberOfHoles.String,
		Private:       row.Private.Bool,
		HostID:        row.HostID.Int32,
	}, nil
}

func (r *EventRepo) GetDetailsByID(ctx context.Context, id int64) (*entity.EventWithDetails, error) {
	row, err := r.q.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.EventWithDetails{
		ID:            row.ID,
		CourseName:    row.CourseName.String,
		Date:          row.Date.String,
		TeeTime:       row.TeeTime.String,
		OpenSpots:     row.OpenSpots.Int32,
		NumberOfHoles: row.NumberOfHoles.String,
		Private:       row.Private.Bool,
		HostName:      row.HostName.String,
		HostID:        row.HostID.Int32,
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
	rows, err := r.q.ListEventsByPlayerID(ctx, sql.NullInt64{Int64: playerID, Valid: true})
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
		RequesterID: sql.NullInt32{Int32: followerID, Valid: true},
		PlayerID:    sql.NullInt64{Int64: playerID, Valid: true},
	})
}

func (r *EventRepo) Create(ctx context.Context, e entity.Event) (int64, error) {
	row, err := r.q.CreateEvent(ctx, sqlcgen.CreateEventParams{
		CourseID:      sql.NullInt32{Int32: e.CourseID, Valid: true},
		Date:          sql.NullString{String: e.Date, Valid: true},
		TeeTime:       sql.NullString{String: e.TeeTime, Valid: true},
		OpenSpots:     sql.NullInt32{Int32: e.OpenSpots, Valid: true},
		NumberOfHoles: sql.NullString{String: e.NumberOfHoles, Valid: true},
		Private:       sql.NullBool{Bool: e.Private, Valid: true},
		HostID:        sql.NullInt32{Int32: e.HostID, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *EventRepo) CreateWithInvites(ctx context.Context, e entity.Event, invitees []int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.q.WithTx(tx)

	event, err := qtx.CreateEvent(ctx, sqlcgen.CreateEventParams{
		CourseID:      sql.NullInt32{Int32: e.CourseID, Valid: true},
		Date:          sql.NullString{String: e.Date, Valid: true},
		TeeTime:       sql.NullString{String: e.TeeTime, Valid: true},
		OpenSpots:     sql.NullInt32{Int32: e.OpenSpots, Valid: true},
		NumberOfHoles: sql.NullString{String: e.NumberOfHoles, Valid: true},
		Private:       sql.NullBool{Bool: e.Private, Valid: true},
		HostID:        sql.NullInt32{Int32: e.HostID, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}

	// Host gets accepted status
	_, err = qtx.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
		PlayerID:     sql.NullInt64{Int64: int64(e.HostID), Valid: true},
		EventID:      sql.NullInt64{Int64: event.ID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: 1, Valid: true}, // accepted
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create host player_event: %w", err)
	}

	if e.Private {
		// Private event: invite only specified invitees
		for _, inviteeID := range invitees {
			if inviteeID == int64(e.HostID) {
				continue
			}
			_, err = qtx.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
				PlayerID:     sql.NullInt64{Int64: inviteeID, Valid: true},
				EventID:      sql.NullInt64{Int64: event.ID, Valid: true},
				InviteStatus: sql.NullInt32{Int32: 0, Valid: true}, // pending
			})
			if err != nil {
				return 0, fmt.Errorf("failed to create invitee player_event: %w", err)
			}
		}
	} else {
		// Public event: invite all players except host
		playerIDs, err := qtx.ListPlayersExceptHost(ctx, int64(e.HostID))
		if err != nil {
			return 0, fmt.Errorf("failed to list players: %w", err)
		}
		for _, pid := range playerIDs {
			_, err = qtx.CreatePlayerEvent(ctx, sqlcgen.CreatePlayerEventParams{
				PlayerID:     sql.NullInt64{Int64: pid, Valid: true},
				EventID:      sql.NullInt64{Int64: event.ID, Valid: true},
				InviteStatus: sql.NullInt32{Int32: 0, Valid: true}, // pending
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
	return r.q.UpdateEvent(ctx, sqlcgen.UpdateEventParams{
		ID:            e.ID,
		CourseID:      sql.NullInt32{Int32: e.CourseID, Valid: true},
		Date:          sql.NullString{String: e.Date, Valid: true},
		TeeTime:       sql.NullString{String: e.TeeTime, Valid: true},
		OpenSpots:     sql.NullInt32{Int32: e.OpenSpots, Valid: true},
		NumberOfHoles: sql.NullString{String: e.NumberOfHoles, Valid: true},
		Private:       sql.NullBool{Bool: e.Private, Valid: true},
	})
}

func (r *EventRepo) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteEvent(ctx, id)
}

func (r *EventRepo) DeletePast(ctx context.Context, today string) error {
	return r.q.DeletePastEvents(ctx, sql.NullString{String: today, Valid: true})
}
