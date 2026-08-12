package booking

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// SearchAvailability resolves the course's provider external id, fetches provider
// slots, upserts tee_times + tee_time_providers, and returns the FindFore cache.
// When minPlayers > 0, only slots with available_slots >= minPlayers (or unknown
// available_slots) are returned.
func (s *Service) SearchAvailability(
	ctx context.Context,
	courseID int64,
	from, to time.Time,
	minPlayers int32,
) ([]entity.TeeTime, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	if courseID == 0 {
		return nil, ErrCourseNotFound
	}
	if !from.Before(to) {
		return nil, fmtInvalidWindow()
	}

	link, err := s.courses.GetProviderByCourse(ctx, courseID, s.provider.ProviderName())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProviderLinkMissing
		}
		return nil, errf("SearchAvailability", err)
	}

	slots, err := s.provider.SearchAvailability(ctx, link.ExternalID, from, to)
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

	listed, err := s.teeTimes.ListByCourseAndWindow(ctx, courseID, from, to)
	if err != nil {
		return nil, errf("SearchAvailability", err)
	}
	if minPlayers <= 0 {
		return listed, nil
	}
	out := make([]entity.TeeTime, 0, len(listed))
	for _, t := range listed {
		if t.AvailableSlots == nil || *t.AvailableSlots >= minPlayers {
			out = append(out, t)
		}
	}
	return out, nil
}

func fmtInvalidWindow() error {
	return errors.New("booking.SearchAvailability: from must be before to")
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
