package product

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a product does not exist.
var ErrNotFound = errors.New("product: not found")

// Repository reads products from storage.
type Repository interface {
	List(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id uint) (Product, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository builds a Repository backed by the given GORM connection.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) List(ctx context.Context) ([]Product, error) {
	var products []Product
	if err := r.db.WithContext(ctx).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("product: list: %w", err)
	}
	return products, nil
}

func (r *gormRepository) GetByID(ctx context.Context, id uint) (Product, error) {
	var p Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("product: get by id: %w", err)
	}
	return p, nil
}
