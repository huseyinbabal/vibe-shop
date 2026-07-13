package order_test

import (
	"context"
	"errors"
	"log"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"

	"vibe-shop/internal/cart"
	"vibe-shop/internal/db"
	"vibe-shop/internal/order"
)

var gormDB *gorm.DB

// Seeded identifiers created in TestMain against a fresh container.
const (
	productA = 1 // Widget, 9.99
	productB = 2 // Gadget, 19.99
	userA    = 1
	userB    = 2
)

const (
	priceA = 9.99
	priceB = 19.99
)

// TestMain spins up a real Postgres container with all five migrations, then
// seeds two products and two users for the package's tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	names := []string{
		"0001_create_products.sql",
		"0002_create_users.sql",
		"0003_create_cart.sql",
		"0004_create_orders.sql",
		"0005_create_order_items.sql",
	}
	migrations := make([]string, 0, len(names))
	for _, name := range names {
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
		"Widget", priceA, "Gadget", priceB).Error; err != nil {
		return err
	}
	return g.Exec(`INSERT INTO users (email, password_hash) VALUES (?, ?), (?, ?)`,
		"a@example.com", "x", "b@example.com", "y").Error
}

func approx(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestCreateFromCart_SnapshotsPricesTotalsAndClearsCart(t *testing.T) {
	carts := cart.NewRepository(gormDB)
	repo := order.NewRepository(gormDB)
	t.Cleanup(func() { _ = carts.ClearByUser(context.Background(), userA) })

	if _, err := carts.AddOrIncrement(context.Background(), userA, productA, 2); err != nil {
		t.Fatalf("add productA: %v", err)
	}
	if _, err := carts.AddOrIncrement(context.Background(), userA, productB, 1); err != nil {
		t.Fatalf("add productB: %v", err)
	}

	placed, err := repo.CreateFromCart(context.Background(), userA)
	if err != nil {
		t.Fatalf("create from cart: %v", err)
	}

	wantTotal := 2*priceA + 1*priceB
	if !approx(placed.Total, wantTotal) {
		t.Errorf("total = %.2f, want %.2f", placed.Total, wantTotal)
	}
	if len(placed.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(placed.Items))
	}
	for _, it := range placed.Items {
		want := priceA
		if it.ProductID == productB {
			want = priceB
		}
		if !approx(it.UnitPrice, want) {
			t.Errorf("product %d unit_price = %.2f, want %.2f", it.ProductID, it.UnitPrice, want)
		}
	}

	// The cart must be emptied by the same transaction.
	lines, err := carts.ListByUser(context.Background(), userA)
	if err != nil {
		t.Fatalf("list cart: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("cart lines after order = %d, want 0", len(lines))
	}

	// Snapshot: a later product price change must not alter the stored order.
	if err := gormDB.Exec(`UPDATE products SET price = ? WHERE id = ?`, 99.99, productA).Error; err != nil {
		t.Fatalf("update price: %v", err)
	}
	t.Cleanup(func() {
		_ = gormDB.Exec(`UPDATE products SET price = ? WHERE id = ?`, priceA, productA).Error
	})

	var stored []order.OrderItem
	if err := gormDB.Where("order_id = ?", placed.ID).Order("product_id").Find(&stored).Error; err != nil {
		t.Fatalf("reload items: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("stored items = %d, want 2", len(stored))
	}
	if !approx(stored[0].UnitPrice, priceA) {
		t.Errorf("stored unit_price = %.2f, want %.2f (price at order time)", stored[0].UnitPrice, priceA)
	}
}

func TestCreateFromCart_EmptyCartReturnsErrCartEmptyAndCreatesNothing(t *testing.T) {
	repo := order.NewRepository(gormDB)

	var before int64
	if err := gormDB.Table("orders").Where("user_id = ?", userB).Count(&before).Error; err != nil {
		t.Fatalf("count orders: %v", err)
	}

	_, err := repo.CreateFromCart(context.Background(), userB)
	if !errors.Is(err, order.ErrCartEmpty) {
		t.Errorf("error = %v, want ErrCartEmpty", err)
	}

	var after int64
	if err := gormDB.Table("orders").Where("user_id = ?", userB).Count(&after).Error; err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if after != before {
		t.Errorf("orders for userB = %d, want %d (no order on empty cart)", after, before)
	}
}

func TestCreateFromCart_IsolatesUsers(t *testing.T) {
	carts := cart.NewRepository(gormDB)
	repo := order.NewRepository(gormDB)
	t.Cleanup(func() {
		_ = carts.ClearByUser(context.Background(), userA)
		_ = carts.ClearByUser(context.Background(), userB)
	})

	if _, err := carts.AddOrIncrement(context.Background(), userA, productA, 1); err != nil {
		t.Fatalf("userA add: %v", err)
	}
	if _, err := carts.AddOrIncrement(context.Background(), userB, productB, 4); err != nil {
		t.Fatalf("userB add: %v", err)
	}

	placed, err := repo.CreateFromCart(context.Background(), userA)
	if err != nil {
		t.Fatalf("userA order: %v", err)
	}
	if placed.UserID != userA {
		t.Errorf("order user_id = %d, want %d", placed.UserID, userA)
	}
	if len(placed.Items) != 1 || placed.Items[0].ProductID != productA {
		t.Errorf("order items = %+v, want only userA's productA line", placed.Items)
	}

	// userB's cart must be untouched by userA's order.
	linesB, err := carts.ListByUser(context.Background(), userB)
	if err != nil {
		t.Fatalf("list userB cart: %v", err)
	}
	if len(linesB) != 1 || linesB[0].Quantity != 4 {
		t.Errorf("userB cart = %+v, want the original single line with quantity 4", linesB)
	}
}
