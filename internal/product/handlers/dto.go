package handlers

import "time"

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
