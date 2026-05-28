package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Stock       int
	Version     int // used for optimistic locking in DecrementStock
	Audit
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
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, limit, offset int) ([]Product, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
	DecrementStock(ctx context.Context, id string, qty, version int) error
}
