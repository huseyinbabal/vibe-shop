package product_test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"

	"vibe-shop/internal/db"
	"vibe-shop/internal/product"
)

var handler *product.Handler

// TestMain spins up a single real Postgres container (via testcontainers-go),
// applies the products migration, and seeds fixed rows for the whole package's
// tests to share — starting a container per test would be needlessly slow.
func TestMain(m *testing.M) {
	ctx := context.Background()

	migrationPath, err := filepath.Abs(filepath.Join("..", "..", "migrations", "0001_create_products.sql"))
	if err != nil {
		log.Fatalf("resolve migration path: %v", err)
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("vibeshop"),
		postgres.WithUsername("vibeshop"),
		postgres.WithPassword("vibeshop"),
		postgres.WithInitScripts(migrationPath),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminate container: %v", err)
		}
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("connection string: %v", err)
	}

	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	if err := seed(gormDB); err != nil {
		log.Fatalf("seed data: %v", err)
	}

	handler = product.NewHandler(product.NewRepository(gormDB))

	os.Exit(m.Run())
}

func seed(gormDB *gorm.DB) error {
	return gormDB.Exec(
		`INSERT INTO products (name, price) VALUES (?, ?), (?, ?)`,
		"Widget", 9.99, "Gadget", 19.99,
	).Error
}

func TestList_ReturnsSeededProducts(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	var products []product.Product
	if err := json.NewDecoder(res.Body).Decode(&products); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(products) != 2 {
		t.Errorf("len(products) = %d, want 2", len(products))
	}
}

func TestGetByID_ReturnsExistingProduct(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	var p product.Product
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if p.Name != "Widget" {
		t.Errorf("Name = %q, want %q", p.Name, "Widget")
	}
}

func TestGetByID_UnknownIDReturns404(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products/9999", nil)
	req.SetPathValue("id", "9999")
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusNotFound)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected non-empty error message in body")
	}
}

func TestGetByID_InvalidIDReturns400(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/products/not-a-number", nil)
	req.SetPathValue("id", "not-a-number")
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", res.StatusCode, http.StatusBadRequest)
	}
}
