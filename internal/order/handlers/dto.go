package handlers

import "time"

// Order Request
type createOrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid4"`
	Quantity  int    `json:"quantity"   validate:"required,gte=1"`
}

type createOrderRequest struct {
	Items []createOrderItemRequest `json:"items"           validate:"required,min=1,dive"`
}

// Order Response
type orderItemResponse struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type orderResponse struct {
	ID          string              `json:"id"`
	UserID      string              `json:"user_id"`
	Items       []orderItemResponse `json:"items"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	CreatedAt   time.Time           `json:"created_at"`
}
