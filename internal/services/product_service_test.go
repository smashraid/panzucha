package services_test

import (
	"context"
	"testing"

	"panzucha/internal/domain"
	"panzucha/internal/services"

	"github.com/google/uuid"
)

type mockProductRepository struct {
	products map[string]*domain.Product
}

func newMockProductRepository() *mockProductRepository {
	return &mockProductRepository{
		products: make(map[string]*domain.Product),
	}
}

func (m *mockProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	// return a copy to mimic DB isolation
	copyProd := *p
	return &copyProd, nil
}

func (m *mockProductRepository) List(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	var list []domain.Product
	for _, p := range m.products {
		list = append(list, *p)
	}
	return list, nil
}

func (m *mockProductRepository) Create(ctx context.Context, p *domain.Product) error {
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepository) Update(ctx context.Context, p *domain.Product) error {
	if _, ok := m.products[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.products[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.products, id)
	return nil
}

func (m *mockProductRepository) DecrementStock(ctx context.Context, id string, qty, version int) error {
	p, ok := m.products[id]
	if !ok {
		return domain.ErrNotFound
	}
	if p.Version != version {
		return domain.ErrVersionConflict
	}
	if p.Stock < qty {
		return domain.ErrInsufficientStock
	}
	p.Stock -= qty
	p.Version++
	return nil
}

func TestProductStockAndOptimisticLocking(t *testing.T) {
	repo := newMockProductRepository()
	svc := services.NewProductService(repo)

	ctx := context.Background()

	product := &domain.Product{
		ID:          uuid.New().String(),
		Name:        "Premium Laptop",
		Description: "A stunning premium notebook",
		Price:       1299.99,
		Stock:       10,
		Version:     1,
	}

	// 1. Create Product
	err := svc.Create(ctx, product)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Retrieve product
	p1, err := svc.GetByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to get product: %v", err)
	}

	// 2. Decrement Stock successfully
	err = svc.DecrementStock(ctx, p1.ID, 3, p1.Version)
	if err != nil {
		t.Fatalf("expected successful stock decrement, got: %v", err)
	}

	// Verify stock and version are updated
	p2, err := svc.GetByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to get product: %v", err)
	}
	if p2.Stock != 7 {
		t.Errorf("expected stock to be 7, got %d", p2.Stock)
	}
	if p2.Version != 2 {
		t.Errorf("expected version to be 2, got %d", p2.Version)
	}

	// 3. Test Version Conflict (Optimistic Locking Failure)
	// Try to decrement using the old version (1) instead of the new version (2)
	err = svc.DecrementStock(ctx, p2.ID, 2, 1)
	if err != domain.ErrVersionConflict {
		t.Errorf("expected ErrVersionConflict, got %v", err)
	}

	// 4. Test Insufficient Stock
	err = svc.DecrementStock(ctx, p2.ID, 10, p2.Version)
	if err != domain.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}
