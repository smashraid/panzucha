package domain

import (
	"context"
	"errors"
	"time"

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
	ID         string      `json:"id" validate:"omitempty,uuid"`
	UserID     string      `json:"user_id" validate:"required,uuid"`
	ProductID  string      `json:"product_id" validate:"required,uuid"`
	Quantity   int         `json:"quantity" validate:"required,gt=0"`
	TotalPrice float64     `json:"total_price" validate:"required,gt=0"`
	Status     OrderStatus `json:"status" validate:"omitempty,oneof=pending paid shipped cancelled"`
	CreatedAt  time.Time   `json:"created_at" validate:"omitempty"`
	UpdatedAt  time.Time   `json:"updated_at" validate:"omitempty"`
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
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByUserID(ctx context.Context, userID string) ([]Order, error)
	UpdateStatus(ctx context.Context, id string, status OrderStatus) error
	List(ctx context.Context) ([]Order, error)
}
