package events

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *PlayerEventService) JoinEvent(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	ev, err := s.events.GetByID(ctx, eventID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil && ev.GroupID != nil {
		if err := requireActiveGroupMember(ctx, s.groups, *ev.GroupID, playerID); err != nil {
			return nil, err
		}
	}
	return s.playerEvents.JoinAccepted(ctx, playerID, eventID)
}
