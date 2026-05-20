package dto

type CreateProductRequest struct {
	Name  string  `json:"name" validate:"required,min=3,max=100"`
	Price float64 `json:"price" validate:"required,gt=0"`
}

type UpdateProductRequest struct {
	Name  string  `json:"name" validate:"omitempty,min=1"`
	Price float64 `json:"price" validate:"omitempty,gt=0"`
}
