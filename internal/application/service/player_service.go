package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// ValidationError represents a validation failure with a user-facing message.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

var emailRegex = regexp.MustCompile(`^[^@\s]+@(?:[-a-z0-9]+\.)+[a-z]{2,}$`)

type PlayerService struct {
	players     port.PlayerRepository
	friendships port.FriendshipRepository
}

func NewPlayerService(players port.PlayerRepository, friendships port.FriendshipRepository) *PlayerService {
	return &PlayerService{players: players, friendships: friendships}
}

func (s *PlayerService) List(ctx context.Context) ([]entity.PlayerWithDetails, error) {
	players, err := s.players.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}

	result := make([]entity.PlayerWithDetails, 0, len(players))
	for _, p := range players {
		details, err := s.GetWithDetails(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("get player details for %d: %w", p.ID, err)
		}
		result = append(result, *details)
	}
	return result, nil
}

func (s *PlayerService) GetWithDetails(ctx context.Context, id int64) (*entity.PlayerWithDetails, error) {
	player, err := s.players.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get player %d: %w", id, err)
	}

	friendIDs, err := s.friendships.ListAcceptedFriendIDs(ctx, int32(player.ID))
	if err != nil {
		return nil, fmt.Errorf("list friend IDs for %d: %w", id, err)
	}

	eventIDs, err := s.friendships.ListAcceptedEventIDs(ctx, player.ID)
	if err != nil {
		return nil, fmt.Errorf("list accepted event IDs for %d: %w", id, err)
	}

	return &entity.PlayerWithDetails{
		ID:       player.ID,
		Name:     player.Name,
		Phone:    player.Phone,
		Email:    player.Email,
		Username: player.Username,
		Friends:  friendIDs,
		Events:   eventIDs,
	}, nil
}

func (s *PlayerService) Create(ctx context.Context, name, phone, email, username, password, passwordConfirmation string) (*entity.Player, error) {
	email = strings.ToLower(email)

	if name == "" {
		return nil, &ValidationError{Message: "Name can't be blank"}
	}
	if phone == "" {
		return nil, &ValidationError{Message: "Phone can't be blank"}
	}
	if email == "" {
		return nil, &ValidationError{Message: "Email can't be blank"}
	}
	if !emailRegex.MatchString(email) {
		return nil, &ValidationError{Message: "Email is invalid"}
	}
	if username == "" {
		return nil, &ValidationError{Message: "Username can't be blank"}
	}
	if password == "" {
		return nil, &ValidationError{Message: "Password can't be blank"}
	}
	if password != passwordConfirmation {
		return nil, &ValidationError{Message: "Password confirmation doesn't match Password"}
	}

	if existing, err := s.players.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, &ValidationError{Message: "Email has already been taken"}
	}
	if existing, err := s.players.GetByUsername(ctx, username); err == nil && existing != nil {
		return nil, &ValidationError{Message: "Username has already been taken"}
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	player, err := s.players.Create(ctx, entity.Player{
		Name:           name,
		Phone:          phone,
		Email:          email,
		Username:       username,
		PasswordDigest: hash,
	})
	if err != nil {
		return nil, fmt.Errorf("create player: %w", err)
	}

	return player, nil
}

func (s *PlayerService) Update(ctx context.Context, callerID int64, name, phone, email, username string) (*entity.PlayerWithDetails, error) {
	email = strings.ToLower(email)

	if name == "" {
		return nil, &ValidationError{Message: "Name can't be blank"}
	}
	if phone == "" {
		return nil, &ValidationError{Message: "Phone can't be blank"}
	}
	if email == "" {
		return nil, &ValidationError{Message: "Email can't be blank"}
	}
	if !emailRegex.MatchString(email) {
		return nil, &ValidationError{Message: "Email is invalid"}
	}
	if username == "" {
		return nil, &ValidationError{Message: "Username can't be blank"}
	}

	if existing, err := s.players.GetByEmail(ctx, email); err == nil && existing != nil && existing.ID != callerID {
		return nil, &ValidationError{Message: "Email has already been taken"}
	}
	if existing, err := s.players.GetByUsername(ctx, username); err == nil && existing != nil && existing.ID != callerID {
		return nil, &ValidationError{Message: "Username has already been taken"}
	}

	_, err := s.players.Update(ctx, entity.Player{
		ID:       callerID,
		Name:     name,
		Phone:    phone,
		Email:    email,
		Username: username,
	})
	if err != nil {
		return nil, fmt.Errorf("update player: %w", err)
	}

	return s.GetWithDetails(ctx, callerID)
}

func (s *PlayerService) ChangePassword(ctx context.Context, callerID int64, currentPassword, newPassword, passwordConfirmation string) error {
	if currentPassword == "" {
		return &ValidationError{Message: "Current password can't be blank"}
	}
	if newPassword == "" {
		return &ValidationError{Message: "New password can't be blank"}
	}
	if newPassword != passwordConfirmation {
		return &ValidationError{Message: "Password confirmation doesn't match"}
	}

	digest, err := s.players.GetPasswordByID(ctx, callerID)
	if err != nil {
		return fmt.Errorf("get password: %w", err)
	}

	if !auth.CheckPassword(currentPassword, digest) {
		return &ValidationError{Message: "Current password is incorrect"}
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.players.UpdatePassword(ctx, callerID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
