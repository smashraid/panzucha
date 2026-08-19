package services_test

import (
	"context"
	"errors"
	"testing"

	"panzucha/internal/order/domain"
	"panzucha/internal/order/services"
	productdomain "panzucha/internal/product/domain"
	shareddomain "panzucha/internal/shared/domain"
	"panzucha/internal/shared/idempotency"
	"panzucha/internal/shared/outbox"

	"github.com/jackc/pgx/v5"
)

// fakeTx satisfies pgx.Tx with only Commit/Rollback implemented —
// the embed trick satisfies the full interface, and the order service
// only ever calls Commit/Rollback on it.
type fakeTx struct {
	pgx.Tx
	committed bool
}

func (f *fakeTx) Commit(context.Context) error   { f.committed = true; return nil }
func (f *fakeTx) Rollback(context.Context) error { return nil }

type fakeTransactor struct {
	err error
}

func (f fakeTransactor) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &fakeTx{}, nil
}

type mockOrderRepository struct {
	orders map[string]*domain.Order
}

func (m *mockOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, shareddomain.ErrNotFound
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

func (m *mockOrderRepository) Create(ctx context.Context, tx pgx.Tx, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	o, ok := m.orders[id]
	if !ok {
		return shareddomain.ErrNotFound
	}
	o.Status = status
	return nil
}

type mockProductRepository struct {
	products map[string]*productdomain.Product
}

func (m *mockProductRepository) GetByID(ctx context.Context, id string) (*productdomain.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, shareddomain.ErrNotFound
	}
	copyProd := *p
	return &copyProd, nil
}

func (m *mockProductRepository) List(ctx context.Context, limit, offset int) ([]productdomain.Product, error) {
	return nil, nil
}
func (m *mockProductRepository) Create(ctx context.Context, p *productdomain.Product) error {
	return nil
}
func (m *mockProductRepository) Update(ctx context.Context, p *productdomain.Product) error {
	return nil
}
func (m *mockProductRepository) Delete(ctx context.Context, id string) error { return nil }

func (m *mockProductRepository) DecrementStock(ctx context.Context, id string, qty, version int) error {
	p, ok := m.products[id]
	if !ok {
		return shareddomain.ErrNotFound
	}
	if p.Version != version {
		return shareddomain.ErrVersionConflict
	}
	if p.Stock < qty {
		return shareddomain.ErrInsufficientStock
	}
	p.Stock -= qty
	p.Version++
	return nil
}

type mockIdempotencyRepository struct {
	keys map[string]*idempotency.IdempotencyKey
}

func (m *mockIdempotencyRepository) FindByKey(ctx context.Context, key string) (*idempotency.IdempotencyKey, error) {
	k, ok := m.keys[key]
	if !ok {
		return nil, shareddomain.ErrNotFound
	}
	return k, nil
}

func (m *mockIdempotencyRepository) Create(ctx context.Context, key *idempotency.IdempotencyKey) error {
	if _, ok := m.keys[key.Key]; ok {
		return shareddomain.ErrConflict
	}
	m.keys[key.Key] = key
	return nil
}

func (m *mockIdempotencyRepository) UpdateToCompleted(ctx context.Context, tx pgx.Tx, key string, resourceID string, statusCode int, responseBody []byte) error {
	k, ok := m.keys[key]
	if !ok {
		return shareddomain.ErrNotFound
	}
	k.Status = idempotency.IdempotencyStatusCompleted
	k.ResourceID = resourceID
	k.ResponseStatus = statusCode
	k.ResponseBody = responseBody
	return nil
}

func (m *mockIdempotencyRepository) Delete(ctx context.Context, key string) error {
	delete(m.keys, key)
	return nil
}

type mockOutboxRepository struct {
	created []outbox.Outbox
}

func (m *mockOutboxRepository) Create(ctx context.Context, tx pgx.Tx, o outbox.Outbox) error {
	m.created = append(m.created, o)
	return nil
}

func (m *mockOutboxRepository) ListAndLock(ctx context.Context, tx pgx.Tx, limit, maxRetries int) ([]outbox.Outbox, error) {
	return nil, nil
}

func (m *mockOutboxRepository) MarkPublishedBatch(ctx context.Context, tx pgx.Tx, ids []string) error {
	return nil
}
func (m *mockOutboxRepository) MarkFailedBatch(ctx context.Context, tx pgx.Tx, ids []string, errMsg string) error {
	return nil
}

func newOrderService(t *testing.T) (services.OrderService, *mockOrderRepository, *mockProductRepository, *mockIdempotencyRepository, *mockOutboxRepository) {
	t.Helper()
	orderRepo := &mockOrderRepository{orders: make(map[string]*domain.Order)}
	productRepo := &mockProductRepository{products: map[string]*productdomain.Product{
		"p1": {ID: "p1", Name: "Laptop", Price: 1000, Stock: 10, Version: 1},
	}}
	idemRepo := &mockIdempotencyRepository{keys: make(map[string]*idempotency.IdempotencyKey)}
	outboxRepo := &mockOutboxRepository{}
	svc := services.NewOrderService(fakeTransactor{}, orderRepo, productRepo, idemRepo, outboxRepo)
	return svc, orderRepo, productRepo, idemRepo, outboxRepo
}

func TestOrderCreate(t *testing.T) {
	svc, orderRepo, _, _, outboxRepo := newOrderService(t)

	order, err := svc.Create(context.Background(), services.CreateOrderInput{
		OrderID:        "o1",
		UserID:         "u1",
		Items:          []domain.OrderItem{{ProductID: "p1", Quantity: 2}},
		IdempotencyKey: "idem-1",
		CreatedBy:      "u1",
	})

	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if order.Status != domain.OrderStatusPending {
		t.Errorf("expected pending status, got %q", order.Status)
	}
	if order.TotalAmount != 2000 {
		t.Errorf("expected total 2000, got %v", order.TotalAmount)
	}
	if _, ok := orderRepo.orders["o1"]; !ok {
		t.Error("expected order to be persisted")
	}
	if len(outboxRepo.created) != 1 {
		t.Errorf("expected 1 outbox row, got %d", len(outboxRepo.created))
	}
}

func TestOrderCreateInsufficientStock(t *testing.T) {
	svc, orderRepo, _, _, outboxRepo := newOrderService(t)

	_, err := svc.Create(context.Background(), services.CreateOrderInput{
		OrderID:        "o1",
		UserID:         "u1",
		Items:          []domain.OrderItem{{ProductID: "p1", Quantity: 99}},
		IdempotencyKey: "idem-1",
		CreatedBy:      "u1",
	})

	if !errors.Is(err, shareddomain.ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
	if _, ok := orderRepo.orders["o1"]; ok {
		t.Error("expected order NOT to be persisted on failure")
	}
	if len(outboxRepo.created) != 0 {
		t.Errorf("expected 0 outbox rows on failure, got %d", len(outboxRepo.created))
	}
}

func TestOrderCreateEmptyItems(t *testing.T) {
	svc, _, _, _, _ := newOrderService(t)

	_, err := svc.Create(context.Background(), services.CreateOrderInput{
		OrderID:        "o1",
		UserID:         "u1",
		Items:          []domain.OrderItem{},
		IdempotencyKey: "idem-1",
		CreatedBy:      "u1",
	})

	if !errors.Is(err, shareddomain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}
