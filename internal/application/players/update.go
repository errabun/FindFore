package players

import (
	"context"
	"fmt"
	"strings"

	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Update(ctx context.Context, callerID int64, name, phone, email, username string) (*entity.PlayerWithDetails, error) {
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

	if existing, err := s.players.GetByEmail(ctx, email); err == nil && existing != nil && existing.ID != callerID {
		return nil, &apperr.ValidationError{Message: "Email has already been taken"}
	}
	if existing, err := s.players.GetByUsername(ctx, username); err == nil && existing != nil && existing.ID != callerID {
		return nil, &apperr.ValidationError{Message: "Username has already been taken"}
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
