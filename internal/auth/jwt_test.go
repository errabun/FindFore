package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret"
	playerID := int64(42)

	token, err := GenerateToken(playerID, 3, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.PlayerID != playerID {
		t.Errorf("ValidateToken returned player_id %d, want %d", claims.PlayerID, playerID)
	}
	if claims.TokenVersion != 3 {
		t.Errorf("ValidateToken returned token_version %d, want 3", claims.TokenVersion)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken(1, 0, "secret-a")
	_, err := ValidateToken(token, "secret-b")
	if err == nil {
		t.Error("ValidateToken should fail with wrong secret")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret"
	claims := jwt.MapClaims{
		"player_id":     float64(1),
		"token_version": float64(0),
		"exp":           time.Now().Add(-1 * time.Hour).Unix(),
		"iat":           time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	_, err := ValidateToken(tokenString, secret)
	if err == nil {
		t.Error("ValidateToken should fail with expired token")
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	_, err := ValidateToken("not-a-valid-token", "secret")
	if err == nil {
		t.Error("ValidateToken should fail with invalid token format")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	if err := ValidatePasswordStrength("short"); err == nil {
		t.Fatal("expected error for short password")
	}
	if err := ValidatePasswordStrength("   "); err == nil {
		t.Fatal("expected error for whitespace password")
	}
	if err := ValidatePasswordStrength("  password1"); err == nil {
		t.Fatal("expected error for leading spaces")
	}
	if err := ValidatePasswordStrength("password1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
