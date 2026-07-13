package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/auth/authtest"
	"vibe-shop/internal/cart"
	"vibe-shop/internal/order"
	"vibe-shop/internal/product"
)

// fakeAuthRepository keeps the router test database-free; only routing is under
// test here, not registration or login behavior.
type fakeAuthRepository struct{}

func (fakeAuthRepository) Create(ctx context.Context, u auth.User) (auth.User, error) {
	return u, nil
}

func (fakeAuthRepository) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	return auth.User{}, auth.ErrNotFound
}

// fakeCartRepository keeps the router test database-free.
type fakeCartRepository struct{}

func (fakeCartRepository) AddOrIncrement(ctx context.Context, userID string, productID uint, quantity int) (cart.Item, error) {
	return cart.Item{}, nil
}

func (fakeCartRepository) ListByUser(ctx context.Context, userID string) ([]cart.LineView, error) {
	return nil, nil
}

func (fakeCartRepository) ClearByUser(ctx context.Context, userID string) error {
	return nil
}

// fakeOrderRepository keeps the router test database-free; ErrCartEmpty lets
// the routing test observe that a request reached the order handler.
type fakeOrderRepository struct{}

func (fakeOrderRepository) CreateFromCart(ctx context.Context, userID string) (order.Order, error) {
	return order.Order{}, order.ErrCartEmpty
}

// testTokens backs the still-routed register/login handler; protected routes
// are guarded by the Keycloak middleware wired in newTestRouter.
var testTokens = auth.NewTokenManager("test-secret", time.Hour)

// newTestRouter wires the router with fake repositories so tests exercise
// routing without a real database. It returns the router and a mint for
// tokens accepted by the Keycloak middleware.
func newTestRouter(t *testing.T) (http.Handler, func(sub string) string) {
	verifier, mint := authtest.New(t)
	products := product.NewHandler(fakeProductRepository{})
	authH := auth.NewHandler(fakeAuthRepository{}, testTokens)
	cartH := cart.NewHandler(fakeCartRepository{})
	ordersH := order.NewHandler(fakeOrderRepository{})
	return NewRouter(products, authH, cartH, ordersH, verifier.RequireAuth), mint
}

// fakeProductRepository is an in-memory stand-in so the router test doesn't
// need a real database — only /health's behavior is under test here.
type fakeProductRepository struct{}

func (fakeProductRepository) List(ctx context.Context) ([]product.Product, error) {
	return nil, nil
}

func (fakeProductRepository) GetByID(ctx context.Context, id uint) (product.Product, error) {
	return product.Product{}, product.ErrNotFound
}

func (fakeProductRepository) Create(ctx context.Context, p product.Product) (product.Product, error) {
	return p, nil
}

func (fakeProductRepository) Update(ctx context.Context, p product.Product) (product.Product, error) {
	return product.Product{}, product.ErrNotFound
}

func (fakeProductRepository) Delete(ctx context.Context, id uint) error {
	return product.ErrNotFound
}

func TestNewRouter_Health(t *testing.T) {
	router, _ := newTestRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	buf := make([]byte, 64)
	n, _ := res.Body.Read(buf)
	if got := strings.TrimSpace(string(buf[:n])); got != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}

// TestNewRouter_OrderRoute proves POST /api/orders is wired behind the auth
// middleware: no token → 401 from the middleware; a valid token reaches the
// handler, whose fake repository answers ErrCartEmpty → 400.
func TestNewRouter_OrderRoute(t *testing.T) {
	router, mint := newTestRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	res, err := srv.Client().Post(srv.URL+"/api/orders", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/orders without token: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("status without token = %d, want 401", res.StatusCode)
	}

	token := mint("router-test-sub")
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/orders", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST /api/orders with token: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status with token = %d, want 400 (fake repo's empty cart)", res.StatusCode)
	}
}

// TestNewRouter_ProductWriteRoutes proves the write routes are wired: each
// request reaches the product handler (status comes from handler logic, not a
// mux-level 404/405). The fake repository keeps the test database-free.
func TestNewRouter_ProductWriteRoutes(t *testing.T) {
	router, _ := newTestRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()

	cases := []struct {
		method, path, body string
		wantStatus         int
	}{
		// Invalid body reaches validation → 400 proves POST is routed.
		{http.MethodPost, "/api/products", `{"name":`, http.StatusBadRequest},
		// Fake repo returns ErrNotFound → 404 proves PUT/DELETE are routed.
		{http.MethodPut, "/api/products/1", `{"name":"x","price":1}`, http.StatusNotFound},
		{http.MethodDelete, "/api/products/1", "", http.StatusNotFound},
	}
	for _, c := range cases {
		t.Run(c.method, func(t *testing.T) {
			req, err := http.NewRequest(c.method, srv.URL+c.path, strings.NewReader(c.body))
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("%s %s: %v", c.method, c.path, err)
			}
			res.Body.Close()
			if res.StatusCode != c.wantStatus {
				t.Errorf("status = %d, want %d", res.StatusCode, c.wantStatus)
			}
		})
	}
}
