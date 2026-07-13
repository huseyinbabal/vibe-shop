package order_test

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
	"vibe-shop/internal/order"
)

var testTokens = auth.NewTokenManager("test-secret", time.Hour)

// authedRequest issues a token for userID and returns a request carrying it.
func authedRequest(t *testing.T, userID uint, method, path string) *http.Request {
	t.Helper()
	token, err := testTokens.Issue(userID)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestCreate_PlacesOrderAndEmptiesCart(t *testing.T) {
	carts := cart.NewRepository(gormDB)
	create := testTokens.RequireAuth(order.NewHandler(order.NewRepository(gormDB)).Create)
	t.Cleanup(func() { _ = carts.ClearByUser(context.Background(), userA) })

	if _, err := carts.AddOrIncrement(context.Background(), userA, productA, 3); err != nil {
		t.Fatalf("add to cart: %v", err)
	}

	rec := httptest.NewRecorder()
	create(rec, authedRequest(t, userA, http.MethodPost, "/api/orders"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	var placed order.Order
	if err := json.Unmarshal(rec.Body.Bytes(), &placed); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if placed.UserID != userA {
		t.Errorf("user_id = %d, want %d", placed.UserID, userA)
	}
	if len(placed.Items) != 1 || placed.Items[0].Quantity != 3 {
		t.Errorf("items = %+v, want one line with quantity 3", placed.Items)
	}
	if !approx(placed.Total, 3*priceA) {
		t.Errorf("total = %.2f, want %.2f", placed.Total, 3*priceA)
	}

	lines, err := carts.ListByUser(context.Background(), userA)
	if err != nil {
		t.Fatalf("list cart: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("cart lines after order = %d, want 0", len(lines))
	}
}

func TestCreate_EmptyCartReturns400(t *testing.T) {
	create := testTokens.RequireAuth(order.NewHandler(order.NewRepository(gormDB)).Create)

	rec := httptest.NewRecorder()
	create(rec, authedRequest(t, userB, http.MethodPost, "/api/orders"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"error":"cart is empty"}` {
		t.Errorf("body = %s, want {\"error\":\"cart is empty\"}", got)
	}
}

func TestCreate_NoTokenReturns401(t *testing.T) {
	create := testTokens.RequireAuth(order.NewHandler(order.NewRepository(gormDB)).Create)

	rec := httptest.NewRecorder()
	create(rec, httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader("")))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
