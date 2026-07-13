package cart_test

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"

	"vibe-shop/internal/cart"
	"vibe-shop/internal/db"
)

var gormDB *gorm.DB

// Seeded identifiers created in TestMain against a fresh container.
const (
	productA = 1
	productB = 2
	userA    = 1
	userB    = 2
)

// TestMain spins up a real Postgres container with the products, users, and
// cart tables, then seeds two products and two users for the package's tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	migrations := make([]string, 0, 3)
	for _, name := range []string{
		"0001_create_products.sql",
		"0002_create_users.sql",
		"0003_create_cart.sql",
	} {
		p, err := filepath.Abs(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			log.Fatalf("resolve migration path %s: %v", name, err)
		}
		migrations = append(migrations, p)
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("vibeshop"),
		postgres.WithUsername("vibeshop"),
		postgres.WithPassword("vibeshop"),
		postgres.WithInitScripts(migrations...),
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

	gormDB, err = db.Connect(dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	if err := seed(gormDB); err != nil {
		log.Fatalf("seed: %v", err)
	}

	os.Exit(m.Run())
}

func seed(g *gorm.DB) error {
	if err := g.Exec(`INSERT INTO products (name, price) VALUES (?, ?), (?, ?)`,
		"Widget", 9.99, "Gadget", 19.99).Error; err != nil {
		return err
	}
	return g.Exec(`INSERT INTO users (email, password_hash) VALUES (?, ?), (?, ?)`,
		"a@example.com", "x", "b@example.com", "y").Error
}

func TestAddOrIncrement_CreatesLine(t *testing.T) {
	repo := cart.NewRepository(gormDB)
	t.Cleanup(func() { _ = repo.ClearByUser(context.Background(), userA) })

	item, err := repo.AddOrIncrement(context.Background(), userA, productA, 2)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if item.Quantity != 2 {
		t.Errorf("quantity = %d, want 2", item.Quantity)
	}
}

func TestAddOrIncrement_SameProductIncrementsInsteadOfNewRow(t *testing.T) {
	repo := cart.NewRepository(gormDB)
	t.Cleanup(func() { _ = repo.ClearByUser(context.Background(), userA) })

	if _, err := repo.AddOrIncrement(context.Background(), userA, productA, 2); err != nil {
		t.Fatalf("first add: %v", err)
	}
	item, err := repo.AddOrIncrement(context.Background(), userA, productA, 3)
	if err != nil {
		t.Fatalf("second add: %v", err)
	}
	if item.Quantity != 5 {
		t.Errorf("quantity = %d, want 5 (2+3)", item.Quantity)
	}

	lines, err := repo.ListByUser(context.Background(), userA)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(lines) != 1 {
		t.Errorf("cart lines = %d, want 1 (incremented, not duplicated)", len(lines))
	}
}

func TestAddOrIncrement_NonExistentProductReturnsErrProductNotFound(t *testing.T) {
	repo := cart.NewRepository(gormDB)

	_, err := repo.AddOrIncrement(context.Background(), userA, 9999, 1)
	if !errors.Is(err, cart.ErrProductNotFound) {
		t.Errorf("error = %v, want ErrProductNotFound", err)
	}
}

func TestListByUser_ReturnsLineTotalsAndIsolatesUsers(t *testing.T) {
	repo := cart.NewRepository(gormDB)
	t.Cleanup(func() {
		_ = repo.ClearByUser(context.Background(), userA)
		_ = repo.ClearByUser(context.Background(), userB)
	})

	if _, err := repo.AddOrIncrement(context.Background(), userA, productA, 3); err != nil {
		t.Fatalf("userA add: %v", err)
	}
	if _, err := repo.AddOrIncrement(context.Background(), userB, productB, 1); err != nil {
		t.Fatalf("userB add: %v", err)
	}

	linesA, err := repo.ListByUser(context.Background(), userA)
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	if len(linesA) != 1 {
		t.Fatalf("userA lines = %d, want 1", len(linesA))
	}
	if linesA[0].ProductID != productA || linesA[0].LineTotal != 3*9.99 {
		t.Errorf("userA line = %+v, want product %d line_total %.2f", linesA[0], productA, 3*9.99)
	}
	// Isolation: userA must not see userB's product.
	for _, l := range linesA {
		if l.ProductID == productB {
			t.Error("userA cart leaked userB's product")
		}
	}
}

func TestClearByUser_EmptiesCart(t *testing.T) {
	repo := cart.NewRepository(gormDB)

	if _, err := repo.AddOrIncrement(context.Background(), userA, productA, 1); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := repo.ClearByUser(context.Background(), userA); err != nil {
		t.Fatalf("clear: %v", err)
	}

	lines, err := repo.ListByUser(context.Background(), userA)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("cart lines after clear = %d, want 0", len(lines))
	}
}
