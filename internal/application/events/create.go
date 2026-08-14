package events

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Create(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error) {
	if e.GroupID != nil {
		if err := requireActiveGroupMember(ctx, s.groups, *e.GroupID, int64(e.HostID)); err != nil {
			return nil, err
		}
		e.Private = true
	}
	if err := s.applyStartsAt(ctx, &e); err != nil {
		return nil, err
	}
	eventID, err := s.events.CreateWithInvites(ctx, e, invitees)
	if err != nil {
		return nil, fmt.Errorf("create event with invites: %w", err)
	}
	return s.buildDetails(ctx, eventID)
}

func (s *Service) applyStartsAt(ctx context.Context, e *entity.Event) error {
	course, err := s.courses.GetByID(ctx, int64(e.CourseID))
	if err != nil {
		return fmt.Errorf("get course %d: %w", e.CourseID, err)
	}
	startsAt, err := ComposeStartsAt(e.Date, e.TeeTime, course.Timezone)
	if err != nil {
		return fmt.Errorf("compose planned_starts_at: %w", err)
	}
	e.PlannedStartsAt = startsAt
	return nil
}
