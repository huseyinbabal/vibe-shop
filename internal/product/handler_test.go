package product_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	// Write tests create and clean up their own rows, so assert the seeded
	// products are present rather than pinning the exact count.
	names := make(map[string]bool, len(products))
	for _, p := range products {
		names[p.Name] = true
	}
	if !names["Widget"] || !names["Gadget"] {
		t.Errorf("products = %v, want Widget and Gadget present", names)
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

// createProduct POSTs a valid product, fails the test on anything but 201, and
// registers a cleanup so tests never leak rows into the shared container.
func createProduct(t *testing.T, name string, price float64) product.Product {
	t.Helper()

	body := fmt.Sprintf(`{"name":%q,"price":%v}`, name, price)
	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d, want %d", res.StatusCode, http.StatusCreated)
	}

	var p product.Product
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		t.Fatalf("create: decode response: %v", err)
	}

	t.Cleanup(func() { deleteProduct(t, p.ID) })
	return p
}

// deleteProduct removes a test-created row; 404 is fine (already deleted).
func deleteProduct(t *testing.T, id uint) {
	t.Helper()

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/products/%d", id), nil)
	req.SetPathValue("id", strconv.FormatUint(uint64(id), 10))
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent && rec.Code != http.StatusNotFound {
		t.Fatalf("cleanup delete: status = %d", rec.Code)
	}
}

func getProduct(t *testing.T, id uint) (int, product.Product) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/products/%d", id), nil)
	req.SetPathValue("id", strconv.FormatUint(uint64(id), 10))
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	var p product.Product
	_ = json.NewDecoder(rec.Result().Body).Decode(&p)
	return rec.Code, p
}

func TestCreate_ValidBodyReturns201AndPersists(t *testing.T) {
	created := createProduct(t, "Lamp", 49.90)

	if created.ID == 0 {
		t.Error("expected non-zero id")
	}
	if created.Name != "Lamp" || math.Abs(created.Price-49.90) > 1e-9 {
		t.Errorf("created = %+v, want name Lamp, price 49.90", created)
	}

	status, got := getProduct(t, created.ID)
	if status != http.StatusOK {
		t.Fatalf("get after create: status = %d, want %d", status, http.StatusOK)
	}
	if got.Name != "Lamp" || math.Abs(got.Price-49.90) > 1e-9 {
		t.Errorf("persisted = %+v, want name Lamp, price 49.90", got)
	}
}

func TestCreate_BodyIDIsIgnored(t *testing.T) {
	body := `{"id":424242,"name":"Vase","price":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusCreated)
	}

	var p product.Product
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	t.Cleanup(func() { deleteProduct(t, p.ID) })

	if p.ID == 424242 {
		t.Error("client-supplied id must be ignored")
	}
}

func TestCreate_InvalidInputReturns400(t *testing.T) {
	cases := map[string]string{
		"invalid JSON":       `{"name":`,
		"empty name":         `{"name":"","price":10}`,
		"whitespace name":    `{"name":"   ","price":10}`,
		"name too long":      fmt.Sprintf(`{"name":%q,"price":10}`, strings.Repeat("a", 201)),
		"zero price":         `{"name":"Lamp","price":0}`,
		"negative price":     `{"name":"Lamp","price":-5}`,
		"too many decimals":  `{"name":"Lamp","price":9.999}`,
		"price above column": `{"name":"Lamp","price":100000000}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Create(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			var errBody map[string]string
			if err := json.NewDecoder(rec.Result().Body).Decode(&errBody); err != nil {
				t.Fatalf("decode error body: %v", err)
			}
			if errBody["error"] == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

func TestUpdate_ExistingReturns200AndPersists(t *testing.T) {
	created := createProduct(t, "Chair", 100)

	body := `{"name":"Comfy Chair","price":149.99}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/products/%d", created.ID), strings.NewReader(body))
	req.SetPathValue("id", strconv.FormatUint(uint64(created.ID), 10))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}

	var p product.Product
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if p.ID != created.ID || p.Name != "Comfy Chair" || math.Abs(p.Price-149.99) > 1e-9 {
		t.Errorf("updated = %+v, want id %d, Comfy Chair, 149.99", p, created.ID)
	}

	status, got := getProduct(t, created.ID)
	if status != http.StatusOK || got.Name != "Comfy Chair" || math.Abs(got.Price-149.99) > 1e-9 {
		t.Errorf("persisted = %+v (status %d), want Comfy Chair, 149.99", got, status)
	}
}

func TestUpdate_InvalidInputReturns400(t *testing.T) {
	created := createProduct(t, "Desk", 200)

	body := `{"name":"","price":10}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/products/%d", created.ID), strings.NewReader(body))
	req.SetPathValue("id", strconv.FormatUint(uint64(created.ID), 10))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdate_UnknownIDReturns404(t *testing.T) {
	body := `{"name":"Ghost","price":10}`
	req := httptest.NewRequest(http.MethodPut, "/api/products/999999", strings.NewReader(body))
	req.SetPathValue("id", "999999")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var errBody map[string]string
	if err := json.NewDecoder(rec.Result().Body).Decode(&errBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Error("expected non-empty error message")
	}
}

func TestUpdate_InvalidIDReturns400(t *testing.T) {
	body := `{"name":"Lamp","price":10}`
	req := httptest.NewRequest(http.MethodPut, "/api/products/not-a-number", strings.NewReader(body))
	req.SetPathValue("id", "not-a-number")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDelete_ExistingReturns204ThenGone(t *testing.T) {
	created := createProduct(t, "Rug", 75.50)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/products/%d", created.ID), nil)
	req.SetPathValue("id", strconv.FormatUint(uint64(created.ID), 10))
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}

	status, _ := getProduct(t, created.ID)
	if status != http.StatusNotFound {
		t.Errorf("get after delete: status = %d, want %d", status, http.StatusNotFound)
	}
}

func TestDelete_UnknownIDReturns404(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/products/999999", nil)
	req.SetPathValue("id", "999999")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDelete_InvalidIDReturns400(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/products/not-a-number", nil)
	req.SetPathValue("id", "not-a-number")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
