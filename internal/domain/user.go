package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string // "admin" | "customer"
	Audit
}

func NewUserID() string {
	return uuid.New().String()
}

func (u *User) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain))
	return err == nil
}

func (u *User) ValidateForCreate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.PasswordHash == "" {
		return errors.New("password is required")
	}
	// optional: email format validation
	return nil
}

func (u *User) ValidateForUpdate() error {
	if u.Name == "" {
		return errors.New("name cannot be empty")
	}
	if u.Email == "" {
		return errors.New("email cannot be empty")
	}
	return nil
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	// List(ctx context.Context) ([]User, error)
}
