package auth_test

import (
	"errors"
	"testing"
	"time"

	"vibe-shop/internal/auth"
)

func TestTokenManager_IssueThenParseRoundTrips(t *testing.T) {
	tm := auth.NewTokenManager("test-secret", time.Hour)

	token, err := tm.Issue(42)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	userID, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if userID != 42 {
		t.Errorf("userID = %d, want 42", userID)
	}
}

func TestTokenManager_ParseExpiredTokenFails(t *testing.T) {
	expired := auth.NewTokenManager("test-secret", -time.Hour)

	token, err := expired.Issue(1)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if _, err := expired.Parse(token); !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("error = %v, want ErrInvalidToken", err)
	}
}

func TestTokenManager_ParseWithWrongSecretFails(t *testing.T) {
	issuer := auth.NewTokenManager("real-secret", time.Hour)
	attacker := auth.NewTokenManager("other-secret", time.Hour)

	token, err := issuer.Issue(7)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if _, err := attacker.Parse(token); !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("error = %v, want ErrInvalidToken", err)
	}
}

func TestTokenManager_ParseMalformedTokenFails(t *testing.T) {
	tm := auth.NewTokenManager("test-secret", time.Hour)

	if _, err := tm.Parse("not-a-jwt"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("error = %v, want ErrInvalidToken", err)
	}
}
