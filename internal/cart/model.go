// Package cart serves a user's shopping cart: one row per product in the cart.
package cart

// Item is a row in the cart table — a single product line in one user's cart.
type Item struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	UserID    uint `gorm:"column:user_id" json:"user_id"`
	ProductID uint `gorm:"column:product_id" json:"product_id"`
	Quantity  int  `json:"quantity"`
}

// TableName pins the GORM table name to the migration's table name.
func (Item) TableName() string {
	return "cart"
}

// LineView is a cart line enriched with product details, used for GET /api/cart.
type LineView struct {
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	LineTotal float64 `json:"line_total"`
}
