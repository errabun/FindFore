package players

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/auth"
)

func (s *Service) ChangePassword(ctx context.Context, callerID int64, currentPassword, newPassword, passwordConfirmation string) error {
	if currentPassword == "" {
		return &apperr.ValidationError{Message: "Current password can't be blank"}
	}
	if err := auth.ValidatePasswordStrength(newPassword); err != nil {
		return &apperr.ValidationError{Message: err.Error()}
	}
	if newPassword != passwordConfirmation {
		return &apperr.ValidationError{Message: "Password confirmation doesn't match"}
	}

	digest, err := s.players.GetPasswordByID(ctx, callerID)
	if err != nil {
		return fmt.Errorf("get password: %w", err)
	}

	if !auth.CheckPassword(currentPassword, digest) {
		return &apperr.ValidationError{Message: "Current password is incorrect"}
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
