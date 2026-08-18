package services

import (
	"context"
	"encoding/json"
	"errors"
	"panzucha/internal/order/domain"
	"panzucha/internal/order/repositories/postgres"
	productdomain "panzucha/internal/product/domain"
	shareddomain "panzucha/internal/shared/domain"
	"panzucha/internal/shared/idempotency"
	"panzucha/internal/shared/outbox"
	"time"

	"github.com/google/uuid"
)

type CreateOrderInput struct {
	OrderID        string
	UserID         string
	Items          []domain.OrderItem
	IdempotencyKey string
	CreatedBy      string
}

type OrderService interface {
	Create(ctx context.Context, req CreateOrderInput) (*domain.Order, error)
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
}

type orderService struct {
	transactor      postgres.Transactor // owned here to begin transactions
	orderRepo       domain.OrderRepository
	productRepo     productdomain.ProductRepository
	idempotencyRepo idempotency.IdempotencyKeyRepository
	outboxRepo      outbox.OutboxRepository
}

func NewOrderService(
	transactor postgres.Transactor,
	orderRepo domain.OrderRepository,
	productRepo productdomain.ProductRepository,
	idempotencyRepo idempotency.IdempotencyKeyRepository,
	outboxRepo outbox.OutboxRepository,
) OrderService {
	return &orderService{
		transactor:      transactor,
		orderRepo:       orderRepo,
		productRepo:     productRepo,
		idempotencyRepo: idempotencyRepo,
		outboxRepo:      outboxRepo,
	}
}

func (s *orderService) Create(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {

	// ── Step 1: Idempotency check ─────────────────────────────────────────
	existing, err := s.idempotencyRepo.FindByKey(ctx, input.IdempotencyKey)
	if err != nil && !errors.Is(err, shareddomain.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		switch existing.Status {
		case idempotency.IdempotencyStatusCompleted:
			return replayOrder(existing)
		case idempotency.IdempotencyStatusProcessing:
			return nil, shareddomain.ErrConflict
		}
	}

	// ── Step 2: Register key as "processing" ──────────────────────────────
	idempotencyEntry := &idempotency.IdempotencyKey{
		Key:          input.IdempotencyKey,
		ResourceType: "order",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	if err := s.idempotencyRepo.Create(ctx, idempotencyEntry); err != nil {
		return nil, err
	}

	// Any failure after this point deletes the key so the client can retry.
	var svcErr error
	defer func() {
		if svcErr != nil {
			_ = s.idempotencyRepo.Delete(ctx, input.IdempotencyKey)
		}
	}()

	// ── Step 3: Validate items and snapshot prices ────────────────────────
	if len(input.Items) == 0 {
		svcErr = shareddomain.ErrInvalidInput
		return nil, svcErr
	}

	total, err := s.snapshotPricesAndTotal(ctx, input.Items)
	if err != nil {
		svcErr = err
		return nil, svcErr
	}

	// ── Step 4: Begin transaction ─────────────────────────────────────────
	// Transactor.BeginTx — no pgxpool in this file.
	tx, err := s.transactor.BeginTx(ctx)
	if err != nil {
		svcErr = err
		return nil, svcErr
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after Commit

	// ── Step 5: Decrement stock for every item (inside tx) ────────────────
	for _, item := range input.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			svcErr = err
			return nil, svcErr
		}
		if err := s.productRepo.DecrementStock(ctx, product.ID, item.Quantity, product.Version); err != nil {
			svcErr = err
			return nil, svcErr
		}
	}

	// ── Step 6: Insert order (inside tx) ─────────────────────────────────
	order := &domain.Order{
		ID:          input.OrderID,
		UserID:      input.UserID,
		Items:       input.Items,
		Status:      domain.OrderStatusPending,
		TotalAmount: total,
		Audit:       shareddomain.Audit{CreatedBy: input.CreatedBy, UpdatedBy: input.CreatedBy},
	}
	if err := s.orderRepo.Create(ctx, tx, order); err != nil {
		svcErr = err
		return nil, svcErr
	}

	// ── Step 7: Mark key completed (inside tx) ────────────────────────────
	responseBody, err := json.Marshal(order)
	if err != nil {
		svcErr = err
		return nil, svcErr
	}
	if err := s.idempotencyRepo.UpdateToCompleted(ctx, tx, input.IdempotencyKey, order.ID, 201, responseBody); err != nil {
		svcErr = err
		return nil, svcErr
	}

	outbox := outbox.Outbox{
		ID:        uuid.NewString(),
		EventID:   uuid.NewString(),
		EventType: domain.EventOrderCreated,
		Payload:   responseBody,
	}
	s.outboxRepo.Create(ctx, tx, outbox)

	// ── Step 8: Commit ────────────────────────────────────────────────────
	if err := tx.Commit(ctx); err != nil {
		svcErr = err
		return nil, svcErr
	}

	return order, nil
}

func (s *orderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *orderService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID, limit, offset)
}

func (s *orderService) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	if err := validateStatusTransition(status); err != nil {
		return err
	}
	return s.orderRepo.UpdateStatus(ctx, id, status)
}

// ── Private helpers ───────────────────────────────────────────────────────────

// snapshotPricesAndTotal fetches the current price for each item from the
// product repo and writes it onto the item. The client sends only product_id
// and quantity — never a price. This prevents price manipulation and ensures
// the persisted order reflects what the customer paid at time of purchase.
func (s *orderService) snapshotPricesAndTotal(ctx context.Context, items []domain.OrderItem) (float64, error) {
	var total float64
	for i, item := range items {
		if item.Quantity <= 0 {
			return 0, shareddomain.ErrInvalidInput
		}
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return 0, err
		}
		items[i].UnitPrice = product.Price
		total += product.Price * float64(item.Quantity)
	}
	return total, nil
}

func replayOrder(key *idempotency.IdempotencyKey) (*domain.Order, error) {
	var o domain.Order
	if err := json.Unmarshal(key.ResponseBody, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func validateStatusTransition(status domain.OrderStatus) error {
	switch status {
	case domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusShipped,
		domain.OrderStatusCancelled:
		return nil
	default:
		return shareddomain.ErrInvalidInput
	}
}
