package events_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type fakeEventRepo struct {
	byID    map[int64]*entity.Event
	details map[int64]*entity.EventWithDetails
}

func newFakeEventRepo() *fakeEventRepo {
	return &fakeEventRepo{
		byID:    make(map[int64]*entity.Event),
		details: make(map[int64]*entity.EventWithDetails),
	}
}

func (r *fakeEventRepo) GetByID(_ context.Context, id int64) (*entity.Event, error) {
	e, ok := r.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *e
	return &cp, nil
}

func (r *fakeEventRepo) GetDetailsByID(_ context.Context, id int64) (*entity.EventWithDetails, error) {
	d, ok := r.details[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *d
	return &cp, nil
}

func (r *fakeEventRepo) ListAllIDs(context.Context) ([]int64, error) { return nil, nil }
func (r *fakeEventRepo) ListPublicIDs(context.Context) ([]int64, error) {
	var ids []int64
	for id, d := range r.details {
		if !d.Private {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
func (r *fakeEventRepo) ListIDsByPlayerID(context.Context, int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeEventRepo) ListFriendsAvailableIDs(context.Context, int32, int64) ([]int64, error) {
	return nil, nil
}
func (r *fakeEventRepo) Create(context.Context, entity.Event) (int64, error) { return 0, nil }
func (r *fakeEventRepo) CreateWithInvites(context.Context, entity.Event, []int64) (int64, error) {
	return 0, nil
}
func (r *fakeEventRepo) Update(context.Context, entity.Event) error { return nil }
func (r *fakeEventRepo) Delete(context.Context, int64) error        { return nil }
func (r *fakeEventRepo) DeletePast(context.Context) error           { return nil }

type fakeCourseRepo struct{}

func (fakeCourseRepo) List(context.Context) ([]entity.Course, error) { return nil, nil }
func (fakeCourseRepo) GetByID(_ context.Context, id int64) (*entity.Course, error) {
	return &entity.Course{ID: id, Timezone: entity.DefaultCourseTimezone}, nil
}
func (fakeCourseRepo) GetByNameAndCity(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (fakeCourseRepo) GetByProviderExternalID(context.Context, string, string) (*entity.Course, error) {
	return nil, sql.ErrNoRows
}
func (fakeCourseRepo) Create(context.Context, entity.Course) (*entity.Course, error) {
	return nil, nil
}
func (fakeCourseRepo) GetProvider(context.Context, string, string) (*entity.CourseProvider, error) {
	return nil, sql.ErrNoRows
}
func (fakeCourseRepo) GetProviderByCourse(context.Context, int64, string) (*entity.CourseProvider, error) {
	return nil, sql.ErrNoRows
}
func (fakeCourseRepo) LinkProvider(context.Context, int64, string, string) error { return nil }

type fakePlayerEventRepo struct {
	byStatus map[int64]map[entity.InviteStatus][]int64
}

func newFakePlayerEventRepo() *fakePlayerEventRepo {
	return &fakePlayerEventRepo{byStatus: make(map[int64]map[entity.InviteStatus][]int64)}
}

func (r *fakePlayerEventRepo) Get(context.Context, int64, int64) (*entity.PlayerEvent, error) {
	return nil, sql.ErrNoRows
}
func (r *fakePlayerEventRepo) Create(context.Context, entity.PlayerEvent) (*entity.PlayerEvent, error) {
	return nil, nil
}
func (r *fakePlayerEventRepo) UpdateStatus(context.Context, int64, int64, entity.InviteStatus) (*entity.PlayerEvent, error) {
	return nil, nil
}
func (r *fakePlayerEventRepo) ListPlayerIDsByEventAndStatus(_ context.Context, eventID int64, status entity.InviteStatus) ([]int64, error) {
	if r.byStatus[eventID] == nil {
		return nil, nil
	}
	return r.byStatus[eventID][status], nil
}
func (r *fakePlayerEventRepo) CountAcceptedForEvent(context.Context, int64) (int64, error) {
	return 0, nil
}
func (r *fakePlayerEventRepo) ClosePendingForEvent(context.Context, int64) error { return nil }
func (r *fakePlayerEventRepo) ReopenClosedForEvent(context.Context, int64) error {
	return nil
}
func (r *fakePlayerEventRepo) JoinAccepted(context.Context, int64, int64) (*entity.PlayerEvent, error) {
	return nil, entity.ErrEventMissing
}
func (r *fakePlayerEventRepo) AcceptInvite(context.Context, int64, int64) (*entity.PlayerEvent, error) {
	return nil, entity.ErrPlayerEventMissing
}

func futureStarts() time.Time {
	loc, _ := time.LoadLocation(entity.DefaultCourseTimezone)
	return time.Date(2099, 1, 1, 8, 0, 0, 0, loc)
}

func TestEventGetPrivateVisibility(t *testing.T) {
	eventRepo := newFakeEventRepo()
	playerEvents := newFakePlayerEventRepo()
	svc := events.NewService(eventRepo, playerEvents, fakeCourseRepo{})

	starts := futureStarts()
	eventRepo.byID[10] = &entity.Event{ID: 10, HostID: 1, Private: true, OpenSpots: 4, PlannedStartsAt: starts}
	eventRepo.details[10] = &entity.EventWithDetails{
		ID: 10, HostID: 1, Private: true, OpenSpots: 4, PlannedStartsAt: starts, CourseTimezone: entity.DefaultCourseTimezone,
	}
	playerEvents.byStatus[10] = map[entity.InviteStatus][]int64{
		entity.InviteStatusPending: {2},
	}

	ctx := context.Background()

	if _, err := svc.Get(ctx, 10, 1); err != nil {
		t.Fatalf("host should view private event: %v", err)
	}
	if _, err := svc.Get(ctx, 10, 2); err != nil {
		t.Fatalf("invitee should view private event: %v", err)
	}
	if _, err := svc.Get(ctx, 10, 3); !errors.Is(err, events.ErrEventNotFound) {
		t.Fatalf("stranger should get not found, got %v", err)
	}
}

func TestEventUpdateDeleteHostOnly(t *testing.T) {
	eventRepo := newFakeEventRepo()
	playerEvents := newFakePlayerEventRepo()
	svc := events.NewService(eventRepo, playerEvents, fakeCourseRepo{})

	starts := futureStarts()
	eventRepo.byID[5] = &entity.Event{ID: 5, HostID: 1, Private: false, OpenSpots: 4, CourseID: 1, PlannedStartsAt: starts}
	eventRepo.details[5] = &entity.EventWithDetails{
		ID: 5, HostID: 1, Private: false, OpenSpots: 4, PlannedStartsAt: starts, CourseTimezone: entity.DefaultCourseTimezone,
	}

	ctx := context.Background()
	_, err := svc.Update(ctx, 2, entity.Event{ID: 5, CourseID: 1, Date: "2099-01-01", TeeTime: "08:00", OpenSpots: 4, NumberOfHoles: "18"}, nil)
	if !errors.Is(err, events.ErrEventForbidden) {
		t.Fatalf("expected forbidden update, got %v", err)
	}
	if err := svc.Delete(ctx, 2, 5); !errors.Is(err, events.ErrEventForbidden) {
		t.Fatalf("expected forbidden delete, got %v", err)
	}
	if err := svc.Delete(ctx, 1, 5); err != nil {
		t.Fatalf("host delete: %v", err)
	}
}

func TestEventListCannotReadAnotherPlayersEvents(t *testing.T) {
	svc := events.NewService(newFakeEventRepo(), newFakePlayerEventRepo(), fakeCourseRepo{})
	other := int64(99)
	_, err := svc.List(context.Background(), 1, &other, false)
	if !errors.Is(err, events.ErrEventForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
