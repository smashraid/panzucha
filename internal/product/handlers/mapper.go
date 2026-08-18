package handlers

import "panzucha/internal/product/domain"

// Product Mapper
func fromCreateProductRequest(req *createProductRequest) *domain.Product {
	return &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}
}

func fromDomainToResponse(p *domain.Product) *productResponse {
	return &productResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
	}
}

// func fromUpdateProductRequest(req *updateProductRequest) *domain.Product {
// 	return &domain.Product{
// 		Name:        req.Name,
// 		Description: req.Description,
// 		Price:       req.Price,
// 		Stock:       req.Stock,
// 	}
// }

func toProductResponse(p domain.Product) productResponse {
	return productResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toProductListResponse(products []domain.Product, limit, offset int) listProductsResponse {
	data := make([]productResponse, len(products))
	for i, p := range products {
		data[i] = toProductResponse(p)
	}
	return listProductsResponse{
		Data:       data,
		TotalCount: len(data),
		Limit:      limit,
		Offset:     offset,
	}
}

func toDomainProduct(req createProductRequest) domain.Product {
	return domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}
}

// Helpers
func applyProductUpdate(existing domain.Product, req updateProductRequest) domain.Product {
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
