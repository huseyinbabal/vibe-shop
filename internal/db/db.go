// Package db establishes the application's database connection.
package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a GORM connection to Postgres using the given DSN.
func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}
	return db, nil
}
