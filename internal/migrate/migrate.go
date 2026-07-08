// Package migrate applies the project's SQL migrations and records which ones
// have run, so the schema is created exactly once and stays in sync on every
// startup without a manual psql step.
package migrate

import (
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// Run applies every *.sql migration in files that has not been recorded yet,
// in filename order, each inside its own transaction. It is safe to call on
// every startup: a migration whose version already exists in schema_migrations
// is skipped, so applied migrations never run twice.
//
// Each file is executed as a single batch. Because Exec is called without bind
// arguments, the pgx driver uses the simple query protocol, which allows a
// migration file to contain more than one statement.
func Run(db *gorm.DB, files fs.FS) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`).Error; err != nil {
		return fmt.Errorf("migrate: ensure tracking table: %w", err)
	}

	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return fmt.Errorf("migrate: list migrations: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")

		var applied int64
		if err := db.Raw(
			"SELECT count(*) FROM schema_migrations WHERE version = ?", version,
		).Scan(&applied).Error; err != nil {
			return fmt.Errorf("migrate: check %s: %w", version, err)
		}
		if applied > 0 {
			continue
		}

		statements, err := fs.ReadFile(files, name)
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", name, err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(statements)).Error; err != nil {
				return err
			}
			return tx.Exec(
				"INSERT INTO schema_migrations (version) VALUES (?)", version,
			).Error
		})
		if err != nil {
			return fmt.Errorf("migrate: apply %s: %w", version, err)
		}
		log.Printf("migrate: applied %s", version)
	}
	return nil
}
