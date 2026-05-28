package services

import (
	"context"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
)

type OrderService interface {
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
	Create(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
}

type orderService struct {
	repo   domain.OrderRepository
	logger *logger.Logger
}

func NewOrderService(repo domain.OrderRepository, log *logger.Logger) OrderService {
	return &orderService{repo: repo, logger: log}
}

func (s *orderService) Create(ctx context.Context, o *domain.Order) error {
	return s.repo.Create(ctx, o)
}

func (s *orderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) GetByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	// Call repo.ListByUser with a high default limit (e.g. 100) and offset 0
	return s.repo.ListByUser(ctx, userID, 100, 0)
}

func (s *orderService) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil || order == nil {
		return err
	}
	if err := order.ValidateForUpdateStatus(status); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *orderService) List(ctx context.Context) ([]domain.Order, error) {
	return s.repo.List(ctx)
}
