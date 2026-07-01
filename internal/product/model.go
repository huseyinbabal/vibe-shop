// Package product serves the product read endpoints.
package product

// Product is a row in the products table.
type Product struct {
	ID    uint    `gorm:"primaryKey" json:"id"`
	Name  string  `json:"name"`
	Price float64 `gorm:"type:numeric(10,2)" json:"price"`
}

// TableName pins the GORM table name to the migration's table name.
func (Product) TableName() string {
	return "products"
}
