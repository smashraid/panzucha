package domain

import (
	"context"
	"panzucha/internal/shared/domain"

	"github.com/jackc/pgx/v5"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

type Order struct {
	ID          string
	UserID      string
	Items       []OrderItem
	Status      OrderStatus
	TotalAmount float64
	domain.Audit
}

type OrderRepository interface {
	GetByID(ctx context.Context, id string) (*Order, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Order, error)
	Create(ctx context.Context, tx pgx.Tx, order *Order) error
	UpdateStatus(ctx context.Context, id string, status OrderStatus) error
}
