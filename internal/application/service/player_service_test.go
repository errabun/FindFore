package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/service"
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

func TestPlayerCreateValidation(t *testing.T) {
	repo := newFakePlayerRepo()
	svc := service.NewPlayerService(repo, newFakeFriendshipRepo())

	tests := []struct {
		name    string
		pass    string
		confirm string
		email   string
		wantMsg string
	}{
		{name: "short password", pass: "short", confirm: "short", email: "a@b.co", wantMsg: "at least 8"},
		{name: "mismatch", pass: "password1", confirm: "password2", email: "a@b.co", wantMsg: "confirmation"},
		{name: "bad email", pass: "password1", confirm: "password1", email: "not-an-email", wantMsg: "Email is invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), "Golfer", "555", tt.email, "golfer1", tt.pass, tt.confirm)
			require.Error(t, err)
			var ve *service.ValidationError
			require.ErrorAs(t, err, &ve)
			assert.Contains(t, ve.Message, tt.wantMsg)
		})
	}
}

func TestPlayerCreateSuccessAndDuplicateEmail(t *testing.T) {
	repo := newFakePlayerRepo()
	svc := service.NewPlayerService(repo, newFakeFriendshipRepo())
	ctx := context.Background()

	p, err := svc.Create(ctx, "Golfer", "555-0100", "golfer@example.com", "golfer1", "password1", "password1")
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "golfer@example.com", p.Email)

	_, err = svc.Create(ctx, "Other", "555-0101", "golfer@example.com", "other1", "password1", "password1")
	var ve *service.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Contains(t, ve.Message, "Email has already been taken")
}

func TestPlayerChangePassword(t *testing.T) {
	repo := newFakePlayerRepo()
	svc := service.NewPlayerService(repo, newFakeFriendshipRepo())
	ctx := context.Background()

	hash, err := auth.HashPassword("oldpass12")
	require.NoError(t, err)
	repo.byID[1] = &entity.Player{ID: 1, PasswordDigest: hash, TokenVersion: 0}

	err = svc.ChangePassword(ctx, 1, "wrongpass", "newpass12", "newpass12")
	var ve *service.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Contains(t, ve.Message, "incorrect")

	require.NoError(t, svc.ChangePassword(ctx, 1, "oldpass12", "newpass12", "newpass12"))
	assert.Equal(t, int32(1), repo.byID[1].TokenVersion)
	assert.True(t, auth.CheckPassword("newpass12", repo.byID[1].PasswordDigest))
}
