package events

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *PlayerEventService) JoinEvent(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	return s.playerEvents.JoinAccepted(ctx, playerID, eventID)
}
