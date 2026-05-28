package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         string
	UserID     string
	ProductID  string
	Quantity   int
	TotalPrice float64
	Status     OrderStatus
	Audit
}

func NewOrderID() string {
	return uuid.New().String()
}

func (o *Order) ValidateForCreate() error {
	if o.Status == "" {
		o.Status = OrderStatusPending
	}
	if o.Status != OrderStatusPending {
		return errors.New("new order status must be 'pending'")
	}
	if o.ProductID == "" {
		return errors.New("product ID is required")
	}
	if o.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	return nil
}

func (o *Order) ValidateForUpdateStatus(newStatus OrderStatus) error {
	switch o.Status {
	case OrderStatusPending:
		if newStatus != OrderStatusPaid && newStatus != OrderStatusCancelled {
			return errors.New("pending order can only become paid or cancelled")
		}
	case OrderStatusPaid:
		if newStatus != OrderStatusShipped {
			return errors.New("paid order can only become shipped")
		}
	case OrderStatusShipped, OrderStatusCancelled:
		return errors.New("order is already final (shipped or cancelled)")
	default:
		return errors.New("unknown order status")
	}
	return nil
}

type OrderRepository interface {
	GetByID(ctx context.Context, id string) (*Order, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Order, error)
	Create(ctx context.Context, order *Order) error
	UpdateStatus(ctx context.Context, id string, status OrderStatus) error
	List(ctx context.Context) ([]Order, error)
}
