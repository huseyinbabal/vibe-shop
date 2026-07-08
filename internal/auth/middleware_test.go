package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"vibe-shop/internal/auth"
)

func TestRequireAuth_RejectsMissingHeader(t *testing.T) {
	tm := auth.NewTokenManager("secret", time.Hour)
	guarded := tm.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not run without a token")
	})

	rec := httptest.NewRecorder()
	guarded(rec, httptest.NewRequest(http.MethodGet, "/api/cart", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAuth_RejectsInvalidToken(t *testing.T) {
	tm := auth.NewTokenManager("secret", time.Hour)
	guarded := tm.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not run with an invalid token")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
	req.Header.Set("Authorization", "Bearer garbage")
	rec := httptest.NewRecorder()
	guarded(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAuth_PassesUserIDInContext(t *testing.T) {
	tm := auth.NewTokenManager("secret", time.Hour)
	token, err := tm.Issue(99)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	var seen uint
	var ran bool
	guarded := tm.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		ran = true
		id, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected a user id in context")
		}
		seen = id
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	guarded(rec, req)

	if !ran {
		t.Fatal("next handler did not run for a valid token")
	}
	if seen != 99 {
		t.Errorf("context user id = %d, want 99", seen)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
