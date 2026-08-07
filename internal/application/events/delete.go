package events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *Service) Delete(ctx context.Context, actorID, id int64) error {
	existing, err := s.events.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEventNotFound
		}
		return fmt.Errorf("get event %d: %w", id, err)
	}
	if int64(existing.HostID) != actorID {
		return ErrEventForbidden
	}
	return s.events.Delete(ctx, id)
}
