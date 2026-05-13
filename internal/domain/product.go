package domain

type Product struct {
	ID    int     `json"id"`
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"gt=0"`
}
