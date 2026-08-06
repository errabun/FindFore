package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type SessionService struct {
	players     port.PlayerRepository
	friendships port.FriendshipRepository
	jwtSecret   string
}

func NewSessionService(players port.PlayerRepository, friendships port.FriendshipRepository, jwtSecret string) *SessionService {
	return &SessionService{
		players:     players,
		friendships: friendships,
		jwtSecret:   jwtSecret,
	}
}

func (s *SessionService) Login(ctx context.Context, login, password string) (*entity.PlayerWithDetails, string, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	var player *entity.Player

	// Try email first, then username
	p, err := s.players.GetByEmail(ctx, login)
	if err == nil {
		player = p
	} else {
		p, err = s.players.GetByUsername(ctx, login)
		if err == nil {
			player = p
		} else {
			return nil, "", fmt.Errorf("invalid email/username or password")
		}
	}

	if !auth.CheckPassword(password, player.PasswordDigest) {
		return nil, "", fmt.Errorf("invalid email/username or password")
	}

	token, err := auth.GenerateToken(player.ID, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	// Build player with details
	friendIDs, err := s.friendships.ListAcceptedFriendIDs(ctx, int32(player.ID))
	if err != nil {
		return nil, "", fmt.Errorf("list friend IDs: %w", err)
	}

	eventIDs, err := s.friendships.ListAcceptedEventIDs(ctx, player.ID)
	if err != nil {
		return nil, "", fmt.Errorf("list accepted event IDs: %w", err)
	}

	details := &entity.PlayerWithDetails{
		ID:       player.ID,
		Name:     player.Name,
		Phone:    player.Phone,
		Email:    player.Email,
		Username: player.Username,
		Friends:  friendIDs,
		Events:   eventIDs,
	}

	return details, token, nil
}
