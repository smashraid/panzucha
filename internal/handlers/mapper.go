package handlers

import "panzucha/internal/domain"

// Order Mapper
// toDomainOrderItems converts the handler's CreateOrderItemRequest slice into
// domain.OrderItem slice. UnitPrice is intentionally left zero — the service
// fetches and snapshots the real price from the product repo. The client
// cannot set or influence the price through this field.
func toDomainOrderItems(reqs []createOrderItemRequest) []domain.OrderItem {
	items := make([]domain.OrderItem, len(reqs))
	for i, r := range reqs {
		items[i] = domain.OrderItem{
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
		}
	}
	return items
}

// ── Outbound: Domain → Response DTO ──────────────────────────────────────────

// toOrderResponse converts a domain.Order into the HTTP response shape.
// CreatedAt comes from the embedded Audit struct.
// Status is cast to string — the client receives the string value of the
// OrderStatus constant, never an internal int or iota.
func toOrderResponse(o domain.Order) orderResponse {
	return orderResponse{
		ID:          o.ID,
		UserID:      o.UserID,
		Items:       toOrderItemResponses(o.Items),
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		CreatedAt:   o.Audit.CreatedAt,
	}
}

// toOrderItemResponses converts a slice of domain.OrderItem into response DTOs.
// UnitPrice is populated here because by the time we reach this mapper the
// service has already snapshotted the price onto each item.
func toOrderItemResponses(items []domain.OrderItem) []orderItemResponse {
	resp := make([]orderItemResponse, len(items))
	for i, item := range items {
		resp[i] = orderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}
	return resp
}

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

// User
func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.Audit.CreatedAt,
		UpdatedAt: u.Audit.UpdatedAt,
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
