package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPasswordAndVerify(t *testing.T) {
	hash, err := HashPassword("StrongPassword123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	matched, err := VerifyPassword("StrongPassword123", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !matched {
		t.Fatal("VerifyPassword() should match the original password")
	}

	matched, err = VerifyPassword("WrongPassword123", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() with wrong password error = %v", err)
	}
	if matched {
		t.Fatal("VerifyPassword() should reject a wrong password")
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("secret", "issuer", 30*time.Minute, 24*time.Hour)
	userID := uuid.New()
	sessionID := uuid.New()

	token, err := manager.GenerateAccessToken(userID, "user", sessionID)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	principal, err := manager.ParseAccessToken(token.Token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	if principal.UserID != userID {
		t.Fatalf("unexpected user id: got %s want %s", principal.UserID, userID)
	}
	if principal.SessionID != sessionID {
		t.Fatalf("unexpected session id: got %s want %s", principal.SessionID, sessionID)
	}
	if principal.Role != "user" {
		t.Fatalf("unexpected role: got %s want user", principal.Role)
	}
}
