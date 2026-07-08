package cart_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/cart"
)

var testTokens = auth.NewTokenManager("test-secret", time.Hour)

// authedRequest issues a token for userID and returns a request carrying it.
func authedRequest(t *testing.T, userID uint, method, path, body string) *http.Request {
	t.Helper()
	token, err := testTokens.Issue(userID)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestAdd_ValidReturns201AndIncrementsOnRepeat(t *testing.T) {
	repo := cart.NewRepository(gormDB)
	h := cart.NewHandler(repo)
	add := testTokens.RequireAuth(h.Add)
	t.Cleanup(func() { _ = repo.ClearByUser(context.Background(), userA) })

	rec := httptest.NewRecorder()
	add(rec, authedRequest(t, userA, http.MethodPost, "/api/cart", `{"product_id":1,"quantity":2}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("first add status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	add(rec, authedRequest(t, userA, http.MethodPost, "/api/cart", `{"product_id":1,"quantity":3}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("second add status = %d, want 201", rec.Code)
	}
	var item cart.Item
	if err := json.Unmarshal(rec.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if item.Quantity != 5 {
		t.Errorf("quantity = %d, want 5 (2+3)", item.Quantity)
	}
}

func TestAdd_InvalidInputAndAuth(t *testing.T) {
	repo := cart.NewRepository(gormDB)
	h := cart.NewHandler(repo)
	add := testTokens.RequireAuth(h.Add)

	t.Run("zero quantity", func(t *testing.T) {
		rec := httptest.NewRecorder()
		add(rec, authedRequest(t, userA, http.MethodPost, "/api/cart", `{"product_id":1,"quantity":0}`))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})
	t.Run("nonexistent product", func(t *testing.T) {
		rec := httptest.NewRecorder()
		add(rec, authedRequest(t, userA, http.MethodPost, "/api/cart", `{"product_id":9999,"quantity":1}`))
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})
	t.Run("invalid json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		add(rec, authedRequest(t, userA, http.MethodPost, "/api/cart", `{"product_id":`))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})
	t.Run("no token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/cart", strings.NewReader(`{"product_id":1,"quantity":1}`))
		add(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})
}

func TestGet_ReturnsTotalsAndIsolatesUsers(t *testing.T) {
	repo := cart.NewRepository(gormDB)
	h := cart.NewHandler(repo)
	add := testTokens.RequireAuth(h.Add)
	get := testTokens.RequireAuth(h.Get)
	t.Cleanup(func() { _ = repo.ClearByUser(context.Background(), userA) })

	rec := httptest.NewRecorder()
	add(rec, authedRequest(t, userA, http.MethodPost, "/api/cart", `{"product_id":1,"quantity":3}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("add status = %d, want 201", rec.Code)
	}

	// userA sees their line and total.
	rec = httptest.NewRecorder()
	get(rec, authedRequest(t, userA, http.MethodGet, "/api/cart", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", rec.Code)
	}
	var body struct {
		Items []cart.LineView `json:"items"`
		Total float64         `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Total != 3*9.99 {
		t.Errorf("items=%d total=%.2f, want 1 item total %.2f", len(body.Items), body.Total, 3*9.99)
	}

	// userB has an empty cart — isolation.
	rec = httptest.NewRecorder()
	get(rec, authedRequest(t, userB, http.MethodGet, "/api/cart", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("userB get status = %d, want 200", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode B: %v", err)
	}
	if len(body.Items) != 0 || body.Total != 0 {
		t.Errorf("userB cart items=%d total=%.2f, want empty", len(body.Items), body.Total)
	}
}
