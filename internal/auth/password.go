package auth

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const MinPasswordLength = 8

// dummyHash is used when no user is found so login timing stays closer across paths.
var dummyHash string

func init() {
	hash, err := bcrypt.GenerateFromPassword([]byte("timing-safe-dummy-password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	dummyHash = string(hash)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CheckPasswordTimingSafe always runs bcrypt. When userFound is false it compares
// against a dummy hash and returns false.
func CheckPasswordTimingSafe(password, hash string, userFound bool) bool {
	if !userFound {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
		return false
	}
	return CheckPassword(password, hash)
}

// ValidatePasswordStrength enforces minimum length and rejects blank/whitespace-only
// or leading/trailing whitespace passwords.
func ValidatePasswordStrength(password string) error {
	if password == "" || strings.TrimSpace(password) == "" {
		return fmt.Errorf("Password can't be blank")
	}
	if password != strings.TrimSpace(password) {
		return fmt.Errorf("Password can't start or end with spaces")
	}
	if len(password) < MinPasswordLength {
		return fmt.Errorf("Password must be at least %d characters", MinPasswordLength)
	}
	return nil
}
