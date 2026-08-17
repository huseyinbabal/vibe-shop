package cart_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"vibe-shop/internal/auth/authtest"
	"vibe-shop/internal/cart"
)

// addOp is one "add this product to the cart" step. Scenarios are expressed as
// a sequence of them so that repeat-add cases sit in the same table as the
// single-add ones.
type addOp struct {
	productID uint
	quantity  int
}

// scenarioUser gives each table row its own Keycloak subject, so rows never
// see each other's cart lines and can be read independently of run order.
func scenarioUser(i int) string {
	return fmt.Sprintf("11111111-aaaa-4bbb-8ccc-1000000000%02d", i)
}

// TestAddOrIncrement_Scenarios drives the repository through the cart's
// add paths and asserts the resulting cart, not just the returned item.
func TestAddOrIncrement_Scenarios(t *testing.T) {
	tests := []struct {
		name string
		adds []addOp
		// wantErr says an add is expected to fail; wantErrIs additionally pins
		// it to a sentinel. A DB-level rejection has no sentinel, so it sets
		// only wantErr.
		wantErr   bool
		wantErrIs error
		wantLines []cart.LineView
	}{
		{
			name:      "empty cart: no adds, no lines",
			adds:      nil,
			wantLines: []cart.LineView{},
		},
		{
			name: "multiple quantity: one line, quantity and line total scale",
			adds: []addOp{{productA, 3}},
			wantLines: []cart.LineView{
				{ProductID: productA, Name: "Widget", Price: 9.99, Quantity: 3, LineTotal: 29.97},
			},
		},
		{
			name:      "nonexistent product: ErrProductNotFound, cart untouched",
			adds:      []addOp{{9999, 1}},
			wantErr:   true,
			wantErrIs: cart.ErrProductNotFound,
			wantLines: []cart.LineView{},
		},
		{
			name: "negative quantity: rejected by the quantity > 0 check, cart untouched",
			adds: []addOp{{productA, -2}},
			// Guarded by the CHECK constraint in 0003_create_cart.sql, which
			// surfaces as a plain wrapped error rather than a sentinel.
			wantErr:   true,
			wantLines: []cart.LineView{},
		},
		{
			name: "same product again: quantities sum onto one line",
			adds: []addOp{{productA, 2}, {productA, 3}},
			wantLines: []cart.LineView{
				{ProductID: productA, Name: "Widget", Price: 9.99, Quantity: 5, LineTotal: 49.95},
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := cart.NewRepository(gormDB)
			userID := scenarioUser(i)
			t.Cleanup(func() { _ = repo.ClearByUser(ctx, userID) })

			var lastErr error
			for _, op := range tt.adds {
				_, lastErr = repo.AddOrIncrement(ctx, userID, op.productID, op.quantity)
				if lastErr != nil {
					break
				}
			}

			switch {
			case tt.wantErr && lastErr == nil:
				t.Fatal("add succeeded, want error")
			case !tt.wantErr && lastErr != nil:
				t.Fatalf("add: %v", lastErr)
			case tt.wantErrIs != nil && !errors.Is(lastErr, tt.wantErrIs):
				t.Fatalf("error = %v, want %v", lastErr, tt.wantErrIs)
			}

			lines, err := repo.ListByUser(ctx, userID)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			assertLines(t, lines, tt.wantLines)
		})
	}
}

// TestCartHTTP_Scenarios runs the same scenarios through the handlers, where
// the status code is the part the client actually reacts to.
func TestCartHTTP_Scenarios(t *testing.T) {
	verifier, mint := authtest.New(t)
	repo := cart.NewRepository(gormDB)
	h := cart.NewHandler(repo)
	add := verifier.RequireAuth(h.Add)
	get := verifier.RequireAuth(h.Get)

	tests := []struct {
		name string
		adds []addOp
		// wantStatus is the status of the last POST; ignored when adds is empty.
		wantStatus int
		wantItems  int
		wantTotal  float64
	}{
		{
			name:      "empty cart: 200 with no items and zero total",
			adds:      nil,
			wantItems: 0,
			wantTotal: 0,
		},
		{
			name:       "multiple quantity: 201 and one line carrying the total",
			adds:       []addOp{{productA, 3}},
			wantStatus: http.StatusCreated,
			wantItems:  1,
			wantTotal:  29.97,
		},
		{
			name:       "nonexistent product: 404 and the cart stays empty",
			adds:       []addOp{{9999, 1}},
			wantStatus: http.StatusNotFound,
			wantItems:  0,
			wantTotal:  0,
		},
		{
			name:       "negative quantity: 400 before reaching the repository",
			adds:       []addOp{{productA, -2}},
			wantStatus: http.StatusBadRequest,
			wantItems:  0,
			wantTotal:  0,
		},
		{
			name:       "same product again: 201 and one merged line",
			adds:       []addOp{{productA, 2}, {productA, 3}},
			wantStatus: http.StatusCreated,
			wantItems:  1,
			wantTotal:  49.95,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Offset the index so these users cannot collide with the
			// repository table's rows.
			userID := scenarioUser(50 + i)
			t.Cleanup(func() { _ = repo.ClearByUser(context.Background(), userID) })

			for j, op := range tt.adds {
				body := fmt.Sprintf(`{"product_id":%d,"quantity":%d}`, op.productID, op.quantity)
				rec := httptest.NewRecorder()
				add(rec, authedRequest(t, mint, userID, http.MethodPost, "/api/cart", body))

				// Only the last POST's status is asserted; earlier ones are
				// setup and must succeed for the scenario to mean anything.
				if j < len(tt.adds)-1 {
					if rec.Code != http.StatusCreated {
						t.Fatalf("setup add %d status = %d, want 201; body=%s", j, rec.Code, rec.Body.String())
					}
					continue
				}
				if rec.Code != tt.wantStatus {
					t.Fatalf("add status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
				}
			}

			rec := httptest.NewRecorder()
			get(rec, authedRequest(t, mint, userID, http.MethodGet, "/api/cart", ""))
			if rec.Code != http.StatusOK {
				t.Fatalf("get status = %d, want 200; body=%s", rec.Code, rec.Body.String())
			}
			var body struct {
				Items []cart.LineView `json:"items"`
				Total float64         `json:"total"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			// items must be [] rather than null, so clients can iterate it.
			if body.Items == nil {
				t.Error("items = null, want []")
			}
			if len(body.Items) != tt.wantItems {
				t.Errorf("items = %d, want %d", len(body.Items), tt.wantItems)
			}
			if body.Total != tt.wantTotal {
				t.Errorf("total = %.2f, want %.2f", body.Total, tt.wantTotal)
			}
		})
	}
}

func assertLines(t *testing.T, got, want []cart.LineView) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("cart lines = %d, want %d (got %+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}
