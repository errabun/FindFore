package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/ericrabun/findfore-go/internal/application/service"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type fakeFriendshipRepo struct {
	nextID int64
	byID   map[int64]*entity.Friendship
}

func newFakeFriendshipRepo() *fakeFriendshipRepo {
	return &fakeFriendshipRepo{
		nextID: 1,
		byID:   make(map[int64]*entity.Friendship),
	}
}

func (r *fakeFriendshipRepo) GetByID(_ context.Context, id int64) (*entity.Friendship, error) {
	f, ok := r.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *f
	return &cp, nil
}

func (r *fakeFriendshipRepo) Find(_ context.Context, requesterID, addresseeID int32) (*entity.Friendship, error) {
	for _, f := range r.byID {
		if f.RequesterID == requesterID && f.AddresseeID == addresseeID {
			cp := *f
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeFriendshipRepo) FindBetween(_ context.Context, playerA, playerB int32) (*entity.Friendship, error) {
	for _, f := range r.byID {
		if (f.RequesterID == playerA && f.AddresseeID == playerB) ||
			(f.RequesterID == playerB && f.AddresseeID == playerA) {
			cp := *f
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeFriendshipRepo) Create(_ context.Context, requesterID, addresseeID int32, status entity.FriendshipStatus) (*entity.Friendship, error) {
	f := &entity.Friendship{
		ID:          r.nextID,
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      status,
	}
	r.nextID++
	r.byID[f.ID] = f
	cp := *f
	return &cp, nil
}

func (r *fakeFriendshipRepo) UpdateStatus(_ context.Context, id int64, status entity.FriendshipStatus) (*entity.Friendship, error) {
	f, ok := r.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	f.Status = status
	cp := *f
	return &cp, nil
}

func (r *fakeFriendshipRepo) DeleteByID(_ context.Context, id int64) error {
	if _, ok := r.byID[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.byID, id)
	return nil
}

func (r *fakeFriendshipRepo) ListAcceptedFriendIDs(_ context.Context, playerID int32) ([]int64, error) {
	var ids []int64
	for _, f := range r.byID {
		if f.Status != entity.FriendshipStatusAccepted {
			continue
		}
		if f.RequesterID == playerID {
			ids = append(ids, int64(f.AddresseeID))
		} else if f.AddresseeID == playerID {
			ids = append(ids, int64(f.RequesterID))
		}
	}
	return ids, nil
}

func (r *fakeFriendshipRepo) ListIncomingPending(_ context.Context, addresseeID int32) ([]entity.Friendship, error) {
	var out []entity.Friendship
	for _, f := range r.byID {
		if f.AddresseeID == addresseeID && f.Status == entity.FriendshipStatusPending {
			out = append(out, *f)
		}
	}
	return out, nil
}

func (r *fakeFriendshipRepo) ListOutgoingPending(_ context.Context, requesterID int32) ([]entity.Friendship, error) {
	var out []entity.Friendship
	for _, f := range r.byID {
		if f.RequesterID == requesterID && f.Status == entity.FriendshipStatusPending {
			out = append(out, *f)
		}
	}
	return out, nil
}

func (r *fakeFriendshipRepo) ListAcceptedEventIDs(_ context.Context, _ int64) ([]int64, error) {
	return nil, nil
}

type fakePlayerSvc struct{}

func (fakePlayerSvc) List(context.Context) ([]entity.PlayerWithDetails, error) {
	return nil, nil
}
func (fakePlayerSvc) GetWithDetails(_ context.Context, id int64) (*entity.PlayerWithDetails, error) {
	return &entity.PlayerWithDetails{ID: id, Name: "Player", Friends: []int64{}, Events: []int64{}}, nil
}
func (fakePlayerSvc) Create(context.Context, string, string, string, string, string, string) (*entity.Player, error) {
	return nil, nil
}
func (fakePlayerSvc) Update(context.Context, int64, string, string, string, string) (*entity.PlayerWithDetails, error) {
	return nil, nil
}
func (fakePlayerSvc) ChangePassword(context.Context, int64, string, string, string) error {
	return nil
}

func TestFriendshipRequestCreatesPending(t *testing.T) {
	repo := newFakeFriendshipRepo()
	svc := service.NewFriendshipService(repo, fakePlayerSvc{})

	f, _, _, err := svc.Request(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if f.Status != entity.FriendshipStatusPending {
		t.Fatalf("expected pending, got %v", f.Status)
	}
	if f.RequesterID != 1 || f.AddresseeID != 2 {
		t.Fatalf("unexpected parties: %+v", f)
	}
}

func TestFriendshipRequestSelfRejected(t *testing.T) {
	svc := service.NewFriendshipService(newFakeFriendshipRepo(), fakePlayerSvc{})
	_, _, _, err := svc.Request(context.Background(), 1, 1)
	if !errors.Is(err, service.ErrFriendshipSelf) {
		t.Fatalf("expected ErrFriendshipSelf, got %v", err)
	}
}

func TestFriendshipAcceptDeclineCancelUnfriend(t *testing.T) {
	repo := newFakeFriendshipRepo()
	svc := service.NewFriendshipService(repo, fakePlayerSvc{})
	ctx := context.Background()

	pending, _, _, err := svc.Request(ctx, 1, 2)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if _, _, _, err := svc.Accept(ctx, 1, pending.ID); !errors.Is(err, service.ErrFriendshipForbidden) {
		t.Fatalf("requester should not accept: %v", err)
	}

	accepted, _, _, err := svc.Accept(ctx, 2, pending.ID)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	if accepted.Status != entity.FriendshipStatusAccepted {
		t.Fatalf("expected accepted, got %v", accepted.Status)
	}

	if err := svc.CancelOrUnfriend(ctx, 2, accepted.ID); err != nil {
		t.Fatalf("unfriend: %v", err)
	}
	if _, err := repo.GetByID(ctx, accepted.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected deleted friendship")
	}

	pending2, _, _, err := svc.Request(ctx, 1, 2)
	if err != nil {
		t.Fatalf("re-request: %v", err)
	}
	if err := svc.Decline(ctx, 2, pending2.ID); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if _, err := repo.GetByID(ctx, pending2.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected declined request deleted")
	}

	// Immediate re-request after decline
	if _, _, _, err := svc.Request(ctx, 1, 2); err != nil {
		t.Fatalf("re-request after decline: %v", err)
	}
}

func TestFriendshipReversePendingAutoAccepts(t *testing.T) {
	repo := newFakeFriendshipRepo()
	svc := service.NewFriendshipService(repo, fakePlayerSvc{})
	ctx := context.Background()

	pending, _, _, err := svc.Request(ctx, 1, 2)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	accepted, _, _, err := svc.Request(ctx, 2, 1)
	if err != nil {
		t.Fatalf("reverse request: %v", err)
	}
	if accepted.ID != pending.ID {
		t.Fatalf("expected same friendship row, got %d vs %d", accepted.ID, pending.ID)
	}
	if accepted.Status != entity.FriendshipStatusAccepted {
		t.Fatalf("expected auto-accepted, got %v", accepted.Status)
	}
}

func TestFriendshipDuplicatePendingRejected(t *testing.T) {
	svc := service.NewFriendshipService(newFakeFriendshipRepo(), fakePlayerSvc{})
	ctx := context.Background()
	if _, _, _, err := svc.Request(ctx, 1, 2); err != nil {
		t.Fatalf("request: %v", err)
	}
	_, _, _, err := svc.Request(ctx, 1, 2)
	if !errors.Is(err, service.ErrFriendshipAlreadyPending) {
		t.Fatalf("expected already pending, got %v", err)
	}
}
