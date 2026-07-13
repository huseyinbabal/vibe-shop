package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vibe-shop/internal/auth/authtest"
	"vibe-shop/internal/cart"
	"vibe-shop/internal/order"
	"vibe-shop/internal/product"
)

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

// fakeProductRepository is an in-memory stand-in so the router test doesn't
// need a real database — only routing is under test here.
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

// newTestRouter wires the router with fake repositories so tests exercise
// routing without a real database. It returns the router and a mint for
// tokens accepted by the Keycloak middleware.
func newTestRouter(t *testing.T) (http.Handler, func(sub string) string) {
	verifier, mint := authtest.New(t)
	products := product.NewHandler(fakeProductRepository{})
	cartH := cart.NewHandler(fakeCartRepository{})
	ordersH := order.NewHandler(fakeOrderRepository{})
	return NewRouter(products, cartH, ordersH, verifier.RequireAuth), mint
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

// TestNewRouter_RegisterLoginRemoved proves slice 5 removed the local auth
// endpoints: user management now lives in Keycloak.
func TestNewRouter_RegisterLoginRemoved(t *testing.T) {
	router, _ := newTestRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	for _, path := range []string{"/api/register", "/api/login"} {
		res, err := srv.Client().Post(srv.URL+path, "application/json", strings.NewReader(`{}`))
		if err != nil {
			t.Fatalf("POST %s: %v", path, err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("POST %s status = %d, want 404", path, res.StatusCode)
		}
	}
}

// TestNewRouter_ProductReadsArePublic proves the read endpoints stay
// reachable without any token.
func TestNewRouter_ProductReadsArePublic(t *testing.T) {
	router, _ := newTestRouter(t)
	srv := httptest.NewServer(router)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/products")
	if err != nil {
		t.Fatalf("GET /api/products: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("GET /api/products status = %d, want 200", res.StatusCode)
	}

	// The fake repository answers ErrNotFound: a 404 (not 401) proves the
	// request reached the handler without a token.
	res, err = http.Get(srv.URL + "/api/products/1")
	if err != nil {
		t.Fatalf("GET /api/products/1: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("GET /api/products/1 status = %d, want 404", res.StatusCode)
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

// TestNewRouter_ProductWriteRoutes proves the write routes are wired behind
// the auth middleware: without a token every write gets 401; with a valid
// token the request reaches the product handler (status comes from handler
// logic, not a mux-level 404/405 or the middleware's 401).
func TestNewRouter_ProductWriteRoutes(t *testing.T) {
	router, mint := newTestRouter(t)
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
		t.Run(c.method+" without token", func(t *testing.T) {
			req, err := http.NewRequest(c.method, srv.URL+c.path, strings.NewReader(c.body))
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("%s %s: %v", c.method, c.path, err)
			}
			res.Body.Close()
			if res.StatusCode != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", res.StatusCode)
			}
		})
		t.Run(c.method+" with token", func(t *testing.T) {
			req, err := http.NewRequest(c.method, srv.URL+c.path, strings.NewReader(c.body))
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+mint("router-test-sub"))
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
