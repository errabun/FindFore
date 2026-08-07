package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/service"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func TestPlayerEventJoinSuccess(t *testing.T) {
	events := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := service.NewPlayerEventService(playerEvents, events)

	events.byID[10] = &entity.Event{ID: 10, HostID: 1, OpenSpots: 2, Private: false}
	playerEvents.acceptedCount[10] = 0

	pe, err := svc.JoinEvent(context.Background(), 5, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(5), pe.PlayerID)
	assert.Equal(t, entity.InviteStatusAccepted, pe.InviteStatus)
	assert.Equal(t, int64(1), playerEvents.acceptedCount[10])
}

func TestPlayerEventJoinRejectsFullAndDuplicate(t *testing.T) {
	events := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := service.NewPlayerEventService(playerEvents, events)
	ctx := context.Background()

	events.byID[10] = &entity.Event{ID: 10, HostID: 1, OpenSpots: 1}
	playerEvents.acceptedCount[10] = 1

	_, err := svc.JoinEvent(ctx, 5, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "full")

	playerEvents.acceptedCount[10] = 0
	playerEvents.existing[playerEventKey{5, 10}] = true
	_, err = svc.JoinEvent(ctx, 5, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already part")
}

func TestPlayerEventUpdateStatusClosesWhenFull(t *testing.T) {
	events := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := service.NewPlayerEventService(playerEvents, events)

	events.byID[3] = &entity.Event{ID: 3, HostID: 1, OpenSpots: 1}
	playerEvents.acceptedCount[3] = 0
	playerEvents.existing[playerEventKey{2, 3}] = true

	_, err := svc.UpdateStatus(context.Background(), 2, 3, "accepted")
	require.NoError(t, err)
	assert.True(t, playerEvents.closed[3], "pending invites should close when full")
}

type playerEventKey struct {
	playerID, eventID int64
}

type joinAwarePlayerEventRepo struct {
	*fakePlayerEventRepo
	existing       map[playerEventKey]bool
	acceptedCount  map[int64]int64
	closed         map[int64]bool
	nextID         int64
}

func newJoinAwarePlayerEventRepo() *joinAwarePlayerEventRepo {
	return &joinAwarePlayerEventRepo{
		fakePlayerEventRepo: newFakePlayerEventRepo(),
		existing:            make(map[playerEventKey]bool),
		acceptedCount:       make(map[int64]int64),
		closed:              make(map[int64]bool),
		nextID:              1,
	}
}

func (r *joinAwarePlayerEventRepo) Get(_ context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	if r.existing[playerEventKey{playerID, eventID}] {
		return &entity.PlayerEvent{PlayerID: playerID, EventID: eventID}, nil
	}
	return nil, sql.ErrNoRows
}

func (r *joinAwarePlayerEventRepo) Create(_ context.Context, pe entity.PlayerEvent) (*entity.PlayerEvent, error) {
	pe.ID = r.nextID
	r.nextID++
	r.existing[playerEventKey{pe.PlayerID, pe.EventID}] = true
	if pe.InviteStatus == entity.InviteStatusAccepted {
		r.acceptedCount[pe.EventID]++
	}
	cp := pe
	return &cp, nil
}

func (r *joinAwarePlayerEventRepo) UpdateStatus(_ context.Context, playerID, eventID int64, status entity.InviteStatus) (*entity.PlayerEvent, error) {
	if status == entity.InviteStatusAccepted {
		r.acceptedCount[eventID]++
	}
	return &entity.PlayerEvent{
		ID:           1,
		PlayerID:     playerID,
		EventID:      eventID,
		InviteStatus: status,
	}, nil
}

func (r *joinAwarePlayerEventRepo) CountAcceptedForEvent(_ context.Context, eventID int64) (int64, error) {
	return r.acceptedCount[eventID], nil
}

func (r *joinAwarePlayerEventRepo) ClosePendingForEvent(_ context.Context, eventID int64) error {
	r.closed[eventID] = true
	return nil
}

func (r *joinAwarePlayerEventRepo) ReopenClosedForEvent(_ context.Context, eventID int64) error {
	r.closed[eventID] = false
	return nil
}
