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

// Product Request
type createProductRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255"`
	Description string  `json:"description" validate:"max=2000"`
	Price       float64 `json:"price"       validate:"required,gt=0"`
	Stock       int     `json:"stock"       validate:"gte=0"`
}

type updateProductRequest struct {
	Name        *string  `json:"name"        validate:"omitempty,min=1,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=2000"`
	Price       *float64 `json:"price"       validate:"omitempty,gt=0"`
	Stock       *int     `json:"stock"       validate:"omitempty,gte=0"`
}

// Product Response
type productResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type listProductsResponse struct {
	Data       []productResponse `json:"data"`
	TotalCount int               `json:"total_count"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// User Request
type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}
type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type updateRequest struct {
	Email string `json:"email" validate:"omitempty,email"`
	Name  string `json:"name" validate:"omitempty"`
}

// User Response
type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
