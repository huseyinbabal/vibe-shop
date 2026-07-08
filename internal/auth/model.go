// Package auth handles user registration, login, and JWT-based request
// authentication.
package auth

import "time"

// User is a row in the users table. PasswordHash is never serialized to
// clients: the json:"-" tag keeps the bcrypt hash out of every response.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName pins the GORM table name to the migration's table name.
func (User) TableName() string {
	return "users"
}
