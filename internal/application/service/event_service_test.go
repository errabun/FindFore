package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/ericrabun/findfore-go/internal/application/service"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type fakeEventRepo struct {
	byID map[int64]*entity.Event
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
func (r *fakeEventRepo) Delete(context.Context, int64) error         { return nil }
func (r *fakeEventRepo) DeletePast(context.Context, string) error     { return nil }

type fakePlayerEventRepo struct {
	// eventID -> status -> player IDs
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
func (r *fakePlayerEventRepo) ClosePendingForEvent(context.Context, int64) error  { return nil }
func (r *fakePlayerEventRepo) ReopenClosedForEvent(context.Context, int64) error { return nil }

func TestEventGetPrivateVisibility(t *testing.T) {
	events := newFakeEventRepo()
	playerEvents := newFakePlayerEventRepo()
	svc := service.NewEventService(events, playerEvents)

	events.byID[10] = &entity.Event{ID: 10, HostID: 1, Private: true, OpenSpots: 4}
	events.details[10] = &entity.EventWithDetails{
		ID: 10, HostID: 1, Private: true, OpenSpots: 4,
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
	if _, err := svc.Get(ctx, 10, 3); !errors.Is(err, service.ErrEventNotFound) {
		t.Fatalf("stranger should get not found, got %v", err)
	}
}

func TestEventUpdateDeleteHostOnly(t *testing.T) {
	events := newFakeEventRepo()
	playerEvents := newFakePlayerEventRepo()
	svc := service.NewEventService(events, playerEvents)

	events.byID[5] = &entity.Event{ID: 5, HostID: 1, Private: false, OpenSpots: 4, CourseID: 1}
	events.details[5] = &entity.EventWithDetails{
		ID: 5, HostID: 1, Private: false, OpenSpots: 4,
	}

	ctx := context.Background()
	_, err := svc.Update(ctx, 2, entity.Event{ID: 5, CourseID: 1, Date: "2099-01-01", TeeTime: "08:00", OpenSpots: 4, NumberOfHoles: "18"}, nil)
	if !errors.Is(err, service.ErrEventForbidden) {
		t.Fatalf("expected forbidden update, got %v", err)
	}
	if err := svc.Delete(ctx, 2, 5); !errors.Is(err, service.ErrEventForbidden) {
		t.Fatalf("expected forbidden delete, got %v", err)
	}
	if err := svc.Delete(ctx, 1, 5); err != nil {
		t.Fatalf("host delete: %v", err)
	}
}

func TestEventListCannotReadAnotherPlayersEvents(t *testing.T) {
	svc := service.NewEventService(newFakeEventRepo(), newFakePlayerEventRepo())
	other := int64(99)
	_, err := svc.List(context.Background(), 1, &other, false)
	if !errors.Is(err, service.ErrEventForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
