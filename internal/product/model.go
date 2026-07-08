// Package product serves the product endpoints.
package product

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

// Product is a row in the products table.
type Product struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Price Money  `gorm:"type:numeric(10,2)" json:"price"`
}

// TableName pins the GORM table name to the migration's table name.
func (Product) TableName() string {
	return "products"
}

const maxNameLength = 200

// maxPrice keeps values inside the NUMERIC(10,2) column so an oversized price
// is a 400 for the client instead of a database error.
var maxPrice = decimal.RequireFromString("99999999.99")

// Input is the client-supplied payload for create and update operations.
// It deliberately has no ID field: the id comes from the path or the database.
type Input struct {
	Name  string `json:"name"`
	Price Money  `json:"price"`
}

// Validate reports the first violated rule, or nil if the input is acceptable.
// POST and PUT share these rules; keep them in this single place.
func (in Input) Validate() error {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return errors.New("name must not be empty")
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return fmt.Errorf("name must be at most %d characters", maxNameLength)
	}
	if in.Price.LessThanOrEqual(decimal.Zero) {
		return errors.New("price must be greater than zero")
	}
	if in.Price.GreaterThan(maxPrice) {
		return fmt.Errorf("price must be at most %s", maxPrice.StringFixed(2))
	}
	if !in.Price.Equal(in.Price.Round(2)) {
		return errors.New("price must have at most two decimal places")
	}
	return nil
}

// product converts validated input into a Product, normalizing the name and
// pinning the price to the column's two-decimal scale.
func (in Input) product(id uint) Product {
	return Product{
		ID:    id,
		Name:  strings.TrimSpace(in.Name),
		Price: Money{in.Price.Round(2)},
	}
}
