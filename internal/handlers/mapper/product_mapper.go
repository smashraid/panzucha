package mapper

import (
	"panzucha/internal/domain"
	"panzucha/internal/handlers/dto"
)

// FromRequestToDomain converts API request to domain model
func FromCreateProductRequest(req *dto.CreateProductRequest) *domain.Product {
	return &domain.Product{
		Name:  req.Name,
		Price: req.Price,
	}
}

// FromDomainToResponse converts domain model to API response
func FromDomainToResponse(p *domain.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:    p.ID,
		Name:  p.Name,
		Price: p.Price,
	}
}

// FromUpdateProductRequest converts update request DTO to domain model.
// Note: ID is not set here – it will be set by the handler from URL param.
func FromUpdateProductRequest(req *dto.UpdateProductRequest) *domain.Product {
	return &domain.Product{
		Name:  req.Name,
		Price: req.Price,
	}
}
