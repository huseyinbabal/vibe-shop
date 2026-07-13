package auth

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ErrNotFound is returned when no user matches the query.
var ErrNotFound = errors.New("auth: user not found")

// ErrEmailTaken is returned when creating a user whose email already exists.
var ErrEmailTaken = errors.New("auth: email already registered")

// Repository reads and writes users in storage.
type Repository interface {
	Create(ctx context.Context, u User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository builds a Repository backed by the given GORM connection.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, u User) (User, error) {
	err := r.db.WithContext(ctx).Create(&u).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return User{}, ErrEmailTaken
	}
	if err != nil {
		return User{}, fmt.Errorf("auth: create user: %w", err)
	}
	return u, nil
}

func (r *gormRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("auth: get user by email: %w", err)
	}
	return u, nil
}
