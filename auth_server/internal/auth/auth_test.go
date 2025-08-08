package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
  "net/http"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "supersecret"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if err := CheckPasswordHash(password, hash); err != nil {
		t.Errorf("CheckPasswordHash() failed for correct password: %v", err)
	}

	if err := CheckPasswordHash("wrongpassword", hash); err == nil {
		t.Error("CheckPasswordHash() succeeded for wrong password, expected error")
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"
	expiresIn := time.Minute

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if gotID != userID {
		t.Errorf("ValidateJWT() got = %v, want %v", gotID, userID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"
	expiresIn := -time.Minute // already expired

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Error("ValidateJWT() succeeded for expired token, expected error")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer 123")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "123" {
		t.Errorf("expected token %q, got %q", "123", token)
	}
}

