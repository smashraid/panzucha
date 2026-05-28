package domain

import (
	"context"
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

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, limit, offset int) ([]Product, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
	DecrementStock(ctx context.Context, id string, qty, version int) error
}
