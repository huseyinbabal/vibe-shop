package order

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ErrCartEmpty is returned when placing an order for a user whose cart is empty.
var ErrCartEmpty = errors.New("order: cart is empty")

// Repository creates orders from a user's cart.
type Repository interface {
	CreateFromCart(ctx context.Context, userID uint) (Order, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository builds a Repository backed by the given GORM connection.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// cartLine is the snapshot of a cart row joined with its product's current price.
type cartLine struct {
	ProductID uint
	Quantity  int
	Price     float64
}

// CreateFromCart converts the user's cart into an order in a single
// transaction: it snapshots each product's current price into order_items,
// records the total, and empties the cart. An empty cart yields ErrCartEmpty
// and no order is created.
func (r *gormRepository) CreateFromCart(ctx context.Context, userID uint) (Order, error) {
	var placed Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lines []cartLine
		if err := tx.
			Table("cart").
			Select("cart.product_id, cart.quantity, products.price").
			Joins("JOIN products ON products.id = cart.product_id").
			Where("cart.user_id = ?", userID).
			Order("cart.product_id").
			Scan(&lines).Error; err != nil {
			return fmt.Errorf("order: load cart: %w", err)
		}
		if len(lines) == 0 {
			return ErrCartEmpty
		}

		var total float64
		items := make([]OrderItem, 0, len(lines))
		for _, l := range lines {
			total += float64(l.Quantity) * l.Price
			items = append(items, OrderItem{
				ProductID: l.ProductID,
				Quantity:  l.Quantity,
				UnitPrice: l.Price,
			})
		}

		placed = Order{UserID: userID, Total: total}
		if err := tx.Create(&placed).Error; err != nil {
			return fmt.Errorf("order: create order: %w", err)
		}

		for i := range items {
			items[i].OrderID = placed.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return fmt.Errorf("order: create items: %w", err)
		}
		placed.Items = items

		if err := tx.Exec("DELETE FROM cart WHERE user_id = ?", userID).Error; err != nil {
			return fmt.Errorf("order: clear cart: %w", err)
		}
		return nil
	})
	if err != nil {
		return Order{}, err
	}
	return placed, nil
}
