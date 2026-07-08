package migrate_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"vibe-shop/internal/db"
	"vibe-shop/internal/migrate"
	"vibe-shop/migrations"
)

// TestRun_AppliesSchemaAndIsIdempotent starts a clean Postgres with no init
// scripts, so the schema exists only if the migrator creates it. It then runs
// the migrator twice to prove applied migrations are not re-run.
func TestRun_AppliesSchemaAndIsIdempotent(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("vibeshop"),
		postgres.WithUsername("vibeshop"),
		postgres.WithPassword("vibeshop"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	gormDB, err := db.Connect(dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	// First run creates the products table and records the migration.
	if err := migrate.Run(gormDB, migrations.FS); err != nil {
		t.Fatalf("first Run: %v", err)
	}

	var hasProducts bool
	if err := gormDB.Raw(
		"SELECT to_regclass('public.products') IS NOT NULL",
	).Scan(&hasProducts).Error; err != nil {
		t.Fatalf("check products table: %v", err)
	}
	if !hasProducts {
		t.Fatal("products table was not created")
	}

	var count int64
	if err := gormDB.Raw("SELECT count(*) FROM schema_migrations").Scan(&count).Error; err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("schema_migrations count = %d, want 1", count)
	}

	// Second run must be a no-op: no error, no duplicate tracking rows.
	if err := migrate.Run(gormDB, migrations.FS); err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if err := gormDB.Raw("SELECT count(*) FROM schema_migrations").Scan(&count).Error; err != nil {
		t.Fatalf("recount schema_migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("after second run count = %d, want 1", count)
	}
}
