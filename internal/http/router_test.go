package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vibe-shop/internal/product"
)

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
	handler := product.NewHandler(fakeProductRepository{})
	srv := httptest.NewServer(NewRouter(handler))
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

// TestNewRouter_ProductWriteRoutes proves the write routes are wired: each
// request reaches the product handler (status comes from handler logic, not a
// mux-level 404/405). The fake repository keeps the test database-free.
func TestNewRouter_ProductWriteRoutes(t *testing.T) {
	handler := product.NewHandler(fakeProductRepository{})
	srv := httptest.NewServer(NewRouter(handler))
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
