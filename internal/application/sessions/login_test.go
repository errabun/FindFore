package sessions_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/sessions"
	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type fakePlayerRepo struct {
	byID       map[int64]*entity.Player
	byEmail    map[string]*entity.Player
	byUsername map[string]*entity.Player
	nextID     int64
}

func newFakePlayerRepo() *fakePlayerRepo {
	return &fakePlayerRepo{
		byID:       make(map[int64]*entity.Player),
		byEmail:    make(map[string]*entity.Player),
		byUsername: make(map[string]*entity.Player),
		nextID:     1,
	}
}

func (r *fakePlayerRepo) List(context.Context) ([]entity.Player, error) {
	out := make([]entity.Player, 0, len(r.byID))
	for _, p := range r.byID {
		out = append(out, *p)
	}
	return out, nil
}

func (r *fakePlayerRepo) GetByID(_ context.Context, id int64) (*entity.Player, error) {
	p, ok := r.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (r *fakePlayerRepo) GetByEmail(_ context.Context, email string) (*entity.Player, error) {
	p, ok := r.byEmail[email]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (r *fakePlayerRepo) GetByUsername(_ context.Context, username string) (*entity.Player, error) {
	p, ok := r.byUsername[username]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (r *fakePlayerRepo) Create(_ context.Context, p entity.Player) (*entity.Player, error) {
	p.ID = r.nextID
	r.nextID++
	cp := p
	r.byID[p.ID] = &cp
	r.byEmail[p.Email] = &cp
	r.byUsername[p.Username] = &cp
	out := cp
	return &out, nil
}

func (r *fakePlayerRepo) Update(_ context.Context, p entity.Player) (*entity.Player, error) {
	existing, ok := r.byID[p.ID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	delete(r.byEmail, existing.Email)
	delete(r.byUsername, existing.Username)
	existing.Name = p.Name
	existing.Phone = p.Phone
	existing.Email = p.Email
	existing.Username = p.Username
	r.byEmail[p.Email] = existing
	r.byUsername[p.Username] = existing
	cp := *existing
	return &cp, nil
}

func (r *fakePlayerRepo) GetPasswordByID(_ context.Context, id int64) (string, error) {
	p, ok := r.byID[id]
	if !ok {
		return "", sql.ErrNoRows
	}
	return p.PasswordDigest, nil
}

func (r *fakePlayerRepo) UpdatePassword(_ context.Context, id int64, passwordDigest string) error {
	p, ok := r.byID[id]
	if !ok {
		return sql.ErrNoRows
	}
	p.PasswordDigest = passwordDigest
	p.TokenVersion++
	return nil
}

func (r *fakePlayerRepo) GetTokenVersion(_ context.Context, id int64) (int32, error) {
	p, ok := r.byID[id]
	if !ok {
		return 0, sql.ErrNoRows
	}
	return p.TokenVersion, nil
}

func (r *fakePlayerRepo) ListIDsExcept(_ context.Context, excludeID int64) ([]int64, error) {
	var ids []int64
	for id := range r.byID {
		if id != excludeID {
			ids = append(ids, id)
		}
	}
	return ids, nil
}


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


func TestSessionLoginSuccess(t *testing.T) {
	players := newFakePlayerRepo()
	friends := newFakeFriendshipRepo()
	svc := sessions.NewService(players, friends, "test-secret")

	hash, err := auth.HashPassword("password1")
	require.NoError(t, err)
	p := &entity.Player{
		ID:             7,
		Name:           "Eric",
		Email:          "eric@example.com",
		Username:       "eric",
		PasswordDigest: hash,
		TokenVersion:   2,
	}
	players.byID[7] = p
	players.byEmail[p.Email] = p
	players.byUsername[p.Username] = p

	details, token, err := svc.Login(context.Background(), "eric@example.com", "password1")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	assert.Equal(t, int64(7), details.ID)

	claims, err := auth.ValidateToken(token, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, int64(7), claims.PlayerID)
	assert.Equal(t, int32(2), claims.TokenVersion)
}

func TestSessionLoginRejectsBadPassword(t *testing.T) {
	players := newFakePlayerRepo()
	svc := sessions.NewService(players, newFakeFriendshipRepo(), "test-secret")

	hash, err := auth.HashPassword("password1")
	require.NoError(t, err)
	p := &entity.Player{ID: 1, Email: "a@b.co", Username: "a", PasswordDigest: hash}
	players.byID[1] = p
	players.byEmail[p.Email] = p

	_, _, err = svc.Login(context.Background(), "a@b.co", "wrong-password")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestSessionLoginRejectsUnknownUser(t *testing.T) {
	svc := sessions.NewService(newFakePlayerRepo(), newFakeFriendshipRepo(), "test-secret")
	_, _, err := svc.Login(context.Background(), "nobody@example.com", "password1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
