package dto

import "time"

type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListProductsResponse struct {
	Data       []ProductResponse `json:"data"`
	TotalCount int               `json:"total_count"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}
