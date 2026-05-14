package dto

type ProductResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int               `json:"total"`
}
