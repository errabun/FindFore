package sessions

import (
	"context"
	"fmt"
	"strings"

	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Login(ctx context.Context, login, password string) (*entity.PlayerWithDetails, string, error) {
	login = strings.ToLower(strings.TrimSpace(login))

	var player *entity.Player
	userFound := false

	if p, err := s.players.GetByEmail(ctx, login); err == nil {
		player = p
		userFound = true
	} else if p, err := s.players.GetByUsername(ctx, login); err == nil {
		player = p
		userFound = true
	}

	digest := ""
	if userFound {
		digest = player.PasswordDigest
	}

	if !auth.CheckPasswordTimingSafe(password, digest, userFound) {
		return nil, "", fmt.Errorf("invalid email/username or password")
	}

	token, err := auth.GenerateToken(player.ID, player.TokenVersion, s.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

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
