package store

import (
	"context"
	"database/sql"
	"fmt"
)

type CreateEventWithInvitesParams struct {
	CourseID      int32
	Date          string
	TeeTime       string
	OpenSpots     int32
	NumberOfHoles string
	Private       bool
	HostID        int32
	Invitees      []int64
}

func CreateEventWithInvites(ctx context.Context, db *sql.DB, q *Queries, params CreateEventWithInvitesParams) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := q.WithTx(tx)

	event, err := qtx.CreateEvent(ctx, CreateEventParams{
		CourseID:      sql.NullInt32{Int32: params.CourseID, Valid: true},
		Date:          sql.NullString{String: params.Date, Valid: true},
		TeeTime:       sql.NullString{String: params.TeeTime, Valid: true},
		OpenSpots:     sql.NullInt32{Int32: params.OpenSpots, Valid: true},
		NumberOfHoles: sql.NullString{String: params.NumberOfHoles, Valid: true},
		Private:       sql.NullBool{Bool: params.Private, Valid: true},
		HostID:        sql.NullInt32{Int32: params.HostID, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}

	// Host gets accepted status
	_, err = qtx.CreatePlayerEvent(ctx, CreatePlayerEventParams{
		PlayerID:     sql.NullInt64{Int64: int64(params.HostID), Valid: true},
		EventID:      sql.NullInt64{Int64: event.ID, Valid: true},
		InviteStatus: sql.NullInt32{Int32: 1, Valid: true}, // accepted
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create host player_event: %w", err)
	}

	if params.Private {
		// Private event: invite only specified invitees
		for _, inviteeID := range params.Invitees {
			if inviteeID == int64(params.HostID) {
				continue
			}
			_, err = qtx.CreatePlayerEvent(ctx, CreatePlayerEventParams{
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
		playerIDs, err := qtx.ListPlayersExceptHost(ctx, int64(params.HostID))
		if err != nil {
			return 0, fmt.Errorf("failed to list players: %w", err)
		}
		for _, pid := range playerIDs {
			_, err = qtx.CreatePlayerEvent(ctx, CreatePlayerEventParams{
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
