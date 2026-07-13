package cart

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrProductNotFound is returned when adding a product that does not exist.
var ErrProductNotFound = errors.New("cart: product not found")

// Repository reads and writes cart lines, always scoped to one user. userID
// is the Keycloak subject taken from the verified token.
type Repository interface {
	AddOrIncrement(ctx context.Context, userID string, productID uint, quantity int) (Item, error)
	ListByUser(ctx context.Context, userID string) ([]LineView, error)
	ClearByUser(ctx context.Context, userID string) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository builds a Repository backed by the given GORM connection.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// AddOrIncrement inserts a cart line, or bumps the quantity of the existing
// line for the same (user, product). It returns the stored line with its total
// quantity.
func (r *gormRepository) AddOrIncrement(ctx context.Context, userID string, productID uint, quantity int) (Item, error) {
	item := Item{UserID: userID, ProductID: productID, Quantity: quantity}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"quantity": gorm.Expr("cart.quantity + ?", quantity),
		}),
	}).Create(&item).Error
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return Item{}, ErrProductNotFound
	}
	if err != nil {
		return Item{}, fmt.Errorf("cart: add or increment: %w", err)
	}

	// On conflict the in-memory item keeps the attempted quantity, not the new
	// total, so re-read the row to report the true stored state.
	var stored Item
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&stored).Error; err != nil {
		return Item{}, fmt.Errorf("cart: reload line: %w", err)
	}
	return stored, nil
}

// ListByUser returns the user's cart lines joined with product name and price,
// each carrying its line total.
func (r *gormRepository) ListByUser(ctx context.Context, userID string) ([]LineView, error) {
	var views []LineView
	err := r.db.WithContext(ctx).
		Table("cart").
		Select("cart.product_id, products.name, products.price, cart.quantity, cart.quantity * products.price AS line_total").
		Joins("JOIN products ON products.id = cart.product_id").
		Where("cart.user_id = ?", userID).
		Order("cart.product_id").
		Scan(&views).Error
	if err != nil {
		return nil, fmt.Errorf("cart: list by user: %w", err)
	}
	return views, nil
}

// ClearByUser removes all of the user's cart lines.
func (r *gormRepository) ClearByUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&Item{}).Error; err != nil {
		return fmt.Errorf("cart: clear by user: %w", err)
	}
	return nil
}
