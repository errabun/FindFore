package booking

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// SearchAvailability fetches provider slots for a course external id, upserts
// tee_times + tee_time_providers (setting last_synced_at), and returns the cache.
func (s *Service) SearchAvailability(
	ctx context.Context,
	courseID int64,
	courseExternalID string,
	from, to time.Time,
) ([]entity.TeeTime, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	slots, err := s.provider.SearchAvailability(ctx, courseExternalID, from, to)
	if err != nil {
		return nil, errf("SearchAvailability", err)
	}

	now := time.Now().UTC()
	provider := s.provider.ProviderName()
	for _, slot := range slots {
		if slot.ExternalID == "" {
			continue
		}
		if _, err := s.upsertSlot(ctx, courseID, provider, slot, now); err != nil {
			return nil, errf("SearchAvailability", err)
		}
	}
	return s.teeTimes.ListByCourseAndWindow(ctx, courseID, from, to)
}

func (s *Service) upsertSlot(ctx context.Context, courseID int64, provider string, slot port.BookingSlot, syncedAt time.Time) (*entity.TeeTime, error) {
	status := slot.Status
	if status == "" {
		status = entity.TeeTimeStatusAvailable
	}

	existing, err := s.teeTimes.GetByProviderExternalID(ctx, provider, slot.ExternalID)
	if err == nil {
		existing.StartsAt = slot.StartsAt
		existing.Holes = slot.Holes
		existing.Status = status
		existing.Capacity = slot.Capacity
		existing.AvailableSlots = slot.AvailableSlots
		existing.PriceCents = slot.PriceCents
		existing.Currency = slot.Currency
		existing.LastSyncedAt = &syncedAt
		return s.teeTimes.UpdateCache(ctx, *existing)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	created, err := s.teeTimes.Create(ctx, entity.TeeTime{
		CourseID:       courseID,
		StartsAt:       slot.StartsAt,
		Holes:          slot.Holes,
		Status:         status,
		Capacity:       slot.Capacity,
		AvailableSlots: slot.AvailableSlots,
		PriceCents:     slot.PriceCents,
		Currency:       slot.Currency,
		LastSyncedAt:   &syncedAt,
	})
	if err != nil {
		return nil, err
	}
	if err := s.teeTimes.LinkProvider(ctx, created.ID, provider, slot.ExternalID); err != nil {
		return nil, err
	}
	return created, nil
}
