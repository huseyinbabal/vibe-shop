package product

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ErrNotFound is returned when a product does not exist.
var ErrNotFound = errors.New("product: not found")

// Repository reads and writes products in storage.
type Repository interface {
	List(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id uint) (Product, error)
	Create(ctx context.Context, p Product) (Product, error)
	Update(ctx context.Context, p Product) (Product, error)
	Delete(ctx context.Context, id uint) error
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

func (r *gormRepository) Create(ctx context.Context, p Product) (Product, error) {
	if err := r.db.WithContext(ctx).Create(&p).Error; err != nil {
		return Product{}, fmt.Errorf("product: create: %w", err)
	}
	return p, nil
}

func (r *gormRepository) Update(ctx context.Context, p Product) (Product, error) {
	res := r.db.WithContext(ctx).Model(&Product{}).Where("id = ?", p.ID).
		Select("name", "price").Updates(Product{Name: p.Name, Price: p.Price})
	if res.Error != nil {
		return Product{}, fmt.Errorf("product: update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return Product{}, ErrNotFound
	}
	return p, nil
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&Product{}, id)
	if res.Error != nil {
		return fmt.Errorf("product: delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
