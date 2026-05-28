package mapper

import (
	"panzucha/internal/domain"
	"panzucha/internal/handlers/dto"
)

func FromCreateProductRequest(req *dto.CreateProductRequest) *domain.Product {
	return &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}
}

func FromDomainToResponse(p *domain.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
	}
}

// func FromUpdateProductRequest(req *dto.UpdateProductRequest) *domain.Product {
// 	return &domain.Product{
// 		Name:        req.Name,
// 		Description: req.Description,
// 		Price:       req.Price,
// 		Stock:       req.Stock,
// 	}
// }

func ToProductResponse(p domain.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ToProductListResponse(products []domain.Product, limit, offset int) dto.ListProductsResponse {
	data := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		data[i] = ToProductResponse(p)
	}
	return dto.ListProductsResponse{
		Data:       data,
		TotalCount: len(data),
		Limit:      limit,
		Offset:     offset,
	}
}

func ToDomainProduct(req dto.CreateProductRequest) domain.Product {
	return domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}
}

func ApplyProductUpdate(existing domain.Product, req dto.UpdateProductRequest) domain.Product {
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.Stock != nil {
		existing.Stock = *req.Stock
	}
	return existing
}
