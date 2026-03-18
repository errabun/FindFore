package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == password {
		t.Error("Hash should not equal the original password")
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}

	if CheckPassword("wrongpassword", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	hash1, _ := HashPassword("password")
	hash2, _ := HashPassword("password")

	if hash1 == hash2 {
		t.Error("Same password should produce different hashes (bcrypt uses random salt)")
	}
}
