package dto

import "panzucha/internal/domain"

type CreateOrderRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	Status domain.OrderStatus `json:"status" validate:"required,oneof=pending paid shipped cancelled"`
}
