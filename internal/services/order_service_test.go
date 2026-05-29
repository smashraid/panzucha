package services_test

import (
	"context"

	"panzucha/internal/domain"
)

type mockOrderRepository struct {
	orders map[string]*domain.Order
}

func newMockOrderRepository() *mockOrderRepository {
	return &mockOrderRepository{
		orders: make(map[string]*domain.Order),
	}
}

func (m *mockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return o, nil
}

func (m *mockOrderRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Order, error) {
	var list []domain.Order
	for _, o := range m.orders {
		if o.UserID == userID {
			list = append(list, *o)
		}
	}
	return list, nil
}

func (m *mockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	o, ok := m.orders[id]
	if !ok {
		return domain.ErrNotFound
	}
	o.Status = status
	return nil
}

func (m *mockOrderRepository) List(ctx context.Context) ([]domain.Order, error) {
	var list []domain.Order
	for _, o := range m.orders {
		list = append(list, *o)
	}
	return list, nil
}

// func TestOrderCreationAndStatusUpdates(t *testing.T) {
// 	repo := newMockOrderRepository()
// 	cfg := &config.Config{
// 		ServiceName: "test-order-service",
// 		Environment: "development",
// 	}
// 	log := logger.New(cfg)
// 	svc := services.NewOrderService(repo)

// 	ctx := context.Background()

// 	order := &svc.CreateOrderInput{
// 		ID:         uuid.NewString(),
// 		UserID:     "user-123",
// 		ProductID:  "product-456",
// 		Quantity:   2,
// 		TotalPrice: 99.98,
// 		Status:     domain.OrderStatusPending,
// 	}

// 	// 1. Create order
// 	err := svc.Create(ctx, order)
// 	if err != nil {
// 		t.Fatalf("failed to create order: %v", err)
// 	}

// 	// 2. Retrieve order and verify
// 	o1, err := svc.GetByID(ctx, order.ID)
// 	if err != nil {
// 		t.Fatalf("failed to get order: %v", err)
// 	}
// 	if o1.Status != domain.OrderStatusPending {
// 		t.Errorf("expected pending status, got %q", o1.Status)
// 	}

// 	// 3. Update order status to paid (legal transition)
// 	err = svc.UpdateStatus(ctx, order.ID, domain.OrderStatusPaid)
// 	if err != nil {
// 		t.Fatalf("failed to update status to paid: %v", err)
// 	}

// 	o2, err := svc.GetByID(ctx, order.ID)
// 	if err != nil {
// 		t.Fatalf("failed to get order: %v", err)
// 	}
// 	if o2.Status != domain.OrderStatusPaid {
// 		t.Errorf("expected paid status, got %q", o2.Status)
// 	}

// 	// 4. Update order status to invalid transition (e.g. back to pending)
// 	err = svc.UpdateStatus(ctx, order.ID, domain.OrderStatusPending)
// 	if err == nil {
// 		t.Error("expected error for illegal status transition (paid -> pending), got nil")
// 	}
// }
