package players

import (
	"context"
	"fmt"
	"strings"

	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Create(ctx context.Context, name, phone, email, username, password, passwordConfirmation string) (*entity.Player, error) {
	email = strings.ToLower(email)

	if name == "" {
		return nil, &apperr.ValidationError{Message: "Name can't be blank"}
	}
	if phone == "" {
		return nil, &apperr.ValidationError{Message: "Phone can't be blank"}
	}
	if email == "" {
		return nil, &apperr.ValidationError{Message: "Email can't be blank"}
	}
	if !emailRegex.MatchString(email) {
		return nil, &apperr.ValidationError{Message: "Email is invalid"}
	}
	if username == "" {
		return nil, &apperr.ValidationError{Message: "Username can't be blank"}
	}
	if err := auth.ValidatePasswordStrength(password); err != nil {
		return nil, &apperr.ValidationError{Message: err.Error()}
	}
	if password != passwordConfirmation {
		return nil, &apperr.ValidationError{Message: "Password confirmation doesn't match Password"}
	}

	if existing, err := s.players.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, &apperr.ValidationError{Message: "Email has already been taken"}
	}
	if existing, err := s.players.GetByUsername(ctx, username); err == nil && existing != nil {
		return nil, &apperr.ValidationError{Message: "Username has already been taken"}
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
