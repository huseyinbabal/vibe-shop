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
