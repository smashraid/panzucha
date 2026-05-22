package mapper

import (
	"panzucha/internal/domain"
	"panzucha/internal/handlers/dto"
)

func FromCreateOrderRequest(req *dto.CreateOrderRequest, userID string) *domain.Order {
	return &domain.Order{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		// TotalPrice will be set by service after fetching product price
		Status: domain.OrderStatusPending,
	}
}

func FromDomainToOrderResponse(o *domain.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:         o.ID,
		UserID:     o.UserID,
		ProductID:  o.ProductID,
		Quantity:   o.Quantity,
		TotalPrice: o.TotalPrice,
		Status:     string(o.Status),
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}
