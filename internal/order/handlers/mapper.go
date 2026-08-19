package handlers

import "panzucha/internal/order/domain"

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
