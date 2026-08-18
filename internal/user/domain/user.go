package domain

import (
	"context"
	"panzucha/internal/shared/domain"
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string // "admin" | "customer"
	domain.Audit
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
