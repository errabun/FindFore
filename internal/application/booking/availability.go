package booking

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// SearchAvailabilityResult is re-exported for callers in this package.
type SearchAvailabilityResult = port.SearchAvailabilityResult

// SearchAvailability resolves the course's provider external id, fetches provider
// slots, upserts tee_times + tee_time_providers, and returns the FindFore cache.
// On provider failure, returns cached rows when present (source=cache); otherwise
// propagates the error. Cached availability is never a booking guarantee.
// When minPlayers > 0, only slots with available_slots >= minPlayers (or unknown
// available_slots) are returned — best-effort filter only.
func (s *Service) SearchAvailability(
	ctx context.Context,
	courseID int64,
	from, to time.Time,
	minPlayers int32,
) (*SearchAvailabilityResult, error) {
	if s.provider == nil {
		return nil, ErrProviderRequired
	}
	if courseID == 0 {
		return nil, ErrCourseNotFound
	}
	if err := validateSearchWindow(from, to, time.Now().UTC()); err != nil {
		return nil, err
	}

	link, err := s.courses.GetProviderByCourse(ctx, courseID, s.provider.ProviderName())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProviderLinkMissing
		}
		return nil, errf("SearchAvailability", err)
	}

	now := time.Now().UTC()
	slots, err := s.provider.SearchAvailability(ctx, link.ExternalID, from, to)
	if err != nil {
		listed, listErr := s.teeTimes.ListByCourseAndWindow(ctx, courseID, from, to)
		if listErr != nil {
			return nil, errf("SearchAvailability", errors.Join(err, listErr))
		}
		if len(listed) == 0 {
			return nil, errf("SearchAvailability", err)
		}
		return &SearchAvailabilityResult{
			TeeTimes:  filterByMinPlayers(listed, minPlayers),
			Source:    port.AvailabilitySourceCache,
			FetchedAt: newestSyncedAt(listed, now),
		}, nil
	}

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
	return &SearchAvailabilityResult{
		TeeTimes:  filterByMinPlayers(listed, minPlayers),
		Source:    port.AvailabilitySourceProvider,
		FetchedAt: now,
	}, nil
}

func validateSearchWindow(from, to, now time.Time) error {
	if !from.Before(to) {
		return ErrInvalidWindow
	}
	if to.Sub(from) > maxSearchWindow {
		return ErrInvalidWindow
	}
	if to.After(now.Add(maxSearchWindow)) {
		return ErrInvalidWindow
	}
	return nil
}

func filterByMinPlayers(listed []entity.TeeTime, minPlayers int32) []entity.TeeTime {
	if minPlayers <= 0 {
		return listed
	}
	out := make([]entity.TeeTime, 0, len(listed))
	for _, t := range listed {
		if t.AvailableSlots == nil || *t.AvailableSlots >= minPlayers {
			out = append(out, t)
		}
	}
	return out
}

func newestSyncedAt(listed []entity.TeeTime, fallback time.Time) time.Time {
	var best time.Time
	for _, t := range listed {
		if t.LastSyncedAt != nil && t.LastSyncedAt.After(best) {
			best = *t.LastSyncedAt
		}
	}
	if best.IsZero() {
		return fallback
	}
	return best
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
