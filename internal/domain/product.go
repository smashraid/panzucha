package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        string    `json:"id"`
	Name      string    `json:"name" validate:"required"`
	Price     float64   `json:"price" validate:"gt=0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewProductID() string {
	return uuid.New().String()
}

func (p *Product) ValidateForCreate() error {
	if p.Name == "" {
		return errors.New("name required")
	}
	if p.Price <= 0 {
		return errors.New("price must be positive")
	}
	// additional create‑only rules
	return nil
}

func (p *Product) ValidateForUpdate() error {
	if p.ID == "" {
		return errors.New("ID required for update")
	}
	if p.Name == "" {
		return errors.New("name cannot be emptied")
	}
	// price can be zero? maybe not allowed to change price? check business rule
	return nil
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]Product, error)
	DecrementStock(ctx context.Context, id string, quantity int) error
}
