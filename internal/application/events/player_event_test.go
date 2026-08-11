package events_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func TestPlayerEventJoinSuccess(t *testing.T) {
	eventRepo := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := events.NewPlayerEventService(playerEvents, eventRepo)

	playerEvents.capacity[10] = 2
	playerEvents.acceptedCount[10] = 0

	pe, err := svc.JoinEvent(context.Background(), 5, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(5), pe.PlayerID)
	assert.Equal(t, entity.InviteStatusAccepted, pe.InviteStatus)
	assert.Equal(t, int64(1), playerEvents.acceptedCount[10])
}

func TestPlayerEventJoinRejectsFullDuplicateAndMissing(t *testing.T) {
	eventRepo := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := events.NewPlayerEventService(playerEvents, eventRepo)
	ctx := context.Background()

	playerEvents.capacity[10] = 1
	playerEvents.acceptedCount[10] = 1

	_, err := svc.JoinEvent(ctx, 5, 10)
	require.ErrorIs(t, err, entity.ErrEventFull)

	playerEvents.acceptedCount[10] = 0
	playerEvents.existing[playerEventKey{5, 10}] = entity.InviteStatusPending
	_, err = svc.JoinEvent(ctx, 5, 10)
	require.ErrorIs(t, err, entity.ErrAlreadyOnEvent)

	_, err = svc.JoinEvent(ctx, 5, 999)
	require.ErrorIs(t, err, entity.ErrEventMissing)
}

func TestPlayerEventAcceptInviteClosesWhenFull(t *testing.T) {
	eventRepo := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := events.NewPlayerEventService(playerEvents, eventRepo)

	playerEvents.capacity[3] = 1
	playerEvents.acceptedCount[3] = 0
	playerEvents.existing[playerEventKey{2, 3}] = entity.InviteStatusPending

	_, err := svc.UpdateStatus(context.Background(), 2, 3, "accepted")
	require.NoError(t, err)
	assert.Equal(t, int64(1), playerEvents.acceptedCount[3])
	assert.True(t, playerEvents.closed[3], "pending invites should close when full")
}

func TestPlayerEventAcceptInviteRejectsWhenFull(t *testing.T) {
	eventRepo := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := events.NewPlayerEventService(playerEvents, eventRepo)

	playerEvents.capacity[3] = 1
	playerEvents.acceptedCount[3] = 1
	playerEvents.existing[playerEventKey{2, 3}] = entity.InviteStatusPending

	_, err := svc.UpdateStatus(context.Background(), 2, 3, "accepted")
	require.ErrorIs(t, err, entity.ErrEventFull)
}

func TestPlayerEventAcceptInviteIdempotentWhenAlreadyAccepted(t *testing.T) {
	eventRepo := newFakeEventRepo()
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := events.NewPlayerEventService(playerEvents, eventRepo)

	playerEvents.capacity[3] = 2
	playerEvents.acceptedCount[3] = 1
	playerEvents.existing[playerEventKey{2, 3}] = entity.InviteStatusAccepted

	pe, err := svc.UpdateStatus(context.Background(), 2, 3, "accepted")
	require.NoError(t, err)
	assert.Equal(t, entity.InviteStatusAccepted, pe.InviteStatus)
	assert.Equal(t, int64(1), playerEvents.acceptedCount[3], "should not double-count")
}

func TestPlayerEventAcceptInviteMissing(t *testing.T) {
	playerEvents := newJoinAwarePlayerEventRepo()
	svc := events.NewPlayerEventService(playerEvents, newFakeEventRepo())
	playerEvents.capacity[3] = 4

	_, err := svc.UpdateStatus(context.Background(), 2, 3, "accepted")
	require.ErrorIs(t, err, entity.ErrPlayerEventMissing)
}

type playerEventKey struct {
	playerID, eventID int64
}

type joinAwarePlayerEventRepo struct {
	*fakePlayerEventRepo
	existing      map[playerEventKey]entity.InviteStatus
	capacity      map[int64]int32
	acceptedCount map[int64]int64
	closed        map[int64]bool
	nextID        int64
}

func newJoinAwarePlayerEventRepo() *joinAwarePlayerEventRepo {
	return &joinAwarePlayerEventRepo{
		fakePlayerEventRepo: newFakePlayerEventRepo(),
		existing:            make(map[playerEventKey]entity.InviteStatus),
		capacity:            make(map[int64]int32),
		acceptedCount:       make(map[int64]int64),
		closed:              make(map[int64]bool),
		nextID:              1,
	}
}

func (r *joinAwarePlayerEventRepo) JoinAccepted(_ context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	capacity, ok := r.capacity[eventID]
	if !ok {
		return nil, entity.ErrEventMissing
	}
	if _, exists := r.existing[playerEventKey{playerID, eventID}]; exists {
		return nil, entity.ErrAlreadyOnEvent
	}
	if r.acceptedCount[eventID] >= int64(capacity) {
		return nil, entity.ErrEventFull
	}
	pe, err := r.Create(context.Background(), entity.PlayerEvent{
		PlayerID:     playerID,
		EventID:      eventID,
		InviteStatus: entity.InviteStatusAccepted,
	})
	if err != nil {
		return nil, err
	}
	if r.acceptedCount[eventID] >= int64(capacity) {
		_ = r.ClosePendingForEvent(context.Background(), eventID)
	}
	return pe, nil
}

func (r *joinAwarePlayerEventRepo) AcceptInvite(_ context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	capacity, ok := r.capacity[eventID]
	if !ok {
		return nil, entity.ErrEventMissing
	}
	status, exists := r.existing[playerEventKey{playerID, eventID}]
	if !exists {
		return nil, entity.ErrPlayerEventMissing
	}
	if status == entity.InviteStatusAccepted {
		return &entity.PlayerEvent{
			ID:           1,
			PlayerID:     playerID,
			EventID:      eventID,
			InviteStatus: entity.InviteStatusAccepted,
		}, nil
	}
	if r.acceptedCount[eventID] >= int64(capacity) {
		return nil, entity.ErrEventFull
	}
	r.existing[playerEventKey{playerID, eventID}] = entity.InviteStatusAccepted
	r.acceptedCount[eventID]++
	if r.acceptedCount[eventID] >= int64(capacity) {
		_ = r.ClosePendingForEvent(context.Background(), eventID)
	}
	return &entity.PlayerEvent{
		ID:           1,
		PlayerID:     playerID,
		EventID:      eventID,
		InviteStatus: entity.InviteStatusAccepted,
	}, nil
}

func (r *joinAwarePlayerEventRepo) Get(_ context.Context, playerID, eventID int64) (*entity.PlayerEvent, error) {
	status, ok := r.existing[playerEventKey{playerID, eventID}]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return &entity.PlayerEvent{PlayerID: playerID, EventID: eventID, InviteStatus: status}, nil
}

func (r *joinAwarePlayerEventRepo) Create(_ context.Context, pe entity.PlayerEvent) (*entity.PlayerEvent, error) {
	pe.ID = r.nextID
	r.nextID++
	r.existing[playerEventKey{pe.PlayerID, pe.EventID}] = pe.InviteStatus
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
	r.existing[playerEventKey{playerID, eventID}] = status
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
