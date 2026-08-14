package events

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func requireActiveGroupMember(ctx context.Context, groups port.GroupRepository, groupID, actorID int64) error {
	if groupID <= 0 || actorID <= 0 || groups == nil {
		return ErrEventNotFound
	}
	if _, err := groups.GetByID(ctx, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEventNotFound
		}
		return err
	}
	m, err := groups.GetMembership(ctx, groupID, actorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEventNotFound
		}
		return err
	}
	if !m.IsActive() {
		return ErrEventNotFound
	}
	return nil
}
