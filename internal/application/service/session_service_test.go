package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/service"
	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func TestSessionLoginSuccess(t *testing.T) {
	players := newFakePlayerRepo()
	friends := newFakeFriendshipRepo()
	svc := service.NewSessionService(players, friends, "test-secret")

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
	svc := service.NewSessionService(players, newFakeFriendshipRepo(), "test-secret")

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
	svc := service.NewSessionService(newFakePlayerRepo(), newFakeFriendshipRepo(), "test-secret")
	_, _, err := svc.Login(context.Background(), "nobody@example.com", "password1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
