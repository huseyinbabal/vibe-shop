// Package order turns a user's cart into a placed order with line items whose
// prices are snapshotted at order time.
package order

import "time"

// Order is a row in the orders table together with its line items.
type Order struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	UserID    uint        `gorm:"column:user_id" json:"user_id"`
	Total     float64     `gorm:"type:numeric(10,2)" json:"total"`
	CreatedAt time.Time   `json:"created_at"`
	Items     []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

// TableName pins the GORM table name to the migration's table name.
func (Order) TableName() string {
	return "orders"
}

// OrderItem is a row in the order_items table. UnitPrice is the product's price
// at the moment the order was placed, so later price changes never alter a
// historical order.
type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   uint    `gorm:"column:order_id" json:"order_id"`
	ProductID uint    `gorm:"column:product_id" json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `gorm:"column:unit_price;type:numeric(10,2)" json:"unit_price"`
}

// TableName pins the GORM table name to the migration's table name.
func (OrderItem) TableName() string {
	return "order_items"
}
