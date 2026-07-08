package auth_test

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

	"vibe-shop/internal/auth"
	"vibe-shop/internal/db"
)

var gormDB *gorm.DB

// TestMain spins up a single real Postgres container, applies the users
// migration, and shares the connection across the package's tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	migration, err := filepath.Abs(filepath.Join("..", "..", "migrations", "0002_create_users.sql"))
	if err != nil {
		log.Fatalf("resolve migration path: %v", err)
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("vibeshop"),
		postgres.WithUsername("vibeshop"),
		postgres.WithPassword("vibeshop"),
		postgres.WithInitScripts(migration),
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

	os.Exit(m.Run())
}

func TestCreate_PersistsUserAndAssignsID(t *testing.T) {
	repo := auth.NewRepository(gormDB)

	created, err := repo.Create(context.Background(), auth.User{
		Email:        "create@example.com",
		PasswordHash: "hashed-value",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected a non-zero id after create")
	}

	got, err := repo.GetByEmail(context.Background(), "create@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if got.ID != created.ID || got.PasswordHash != "hashed-value" {
		t.Errorf("round-trip mismatch: got %+v, created %+v", got, created)
	}
}

func TestCreate_DuplicateEmailReturnsErrEmailTaken(t *testing.T) {
	repo := auth.NewRepository(gormDB)

	_, err := repo.Create(context.Background(), auth.User{
		Email:        "dupe@example.com",
		PasswordHash: "h1",
	})
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = repo.Create(context.Background(), auth.User{
		Email:        "dupe@example.com",
		PasswordHash: "h2",
	})
	if !errors.Is(err, auth.ErrEmailTaken) {
		t.Errorf("second create error = %v, want ErrEmailTaken", err)
	}
}

func TestGetByEmail_MissingReturnsErrNotFound(t *testing.T) {
	repo := auth.NewRepository(gormDB)

	_, err := repo.GetByEmail(context.Background(), "nobody@example.com")
	if !errors.Is(err, auth.ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}
