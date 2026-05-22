package services

import (
	"context"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
)

type OrderService interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	List(ctx context.Context) ([]domain.Order, error)
}

type orderService struct {
	repo   domain.OrderRepository
	logger *logger.Logger
}

// Fixed constructor – now accepts domain.OrderRepository
func NewOrderService(repo domain.OrderRepository, log *logger.Logger) OrderService {
	return &orderService{repo: repo, logger: log}
}

func (s *orderService) Create(ctx context.Context, o *domain.Order) error {
	if err := o.ValidateForCreate(); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityCreate,
			EntityType:  "order",
			EntityID:    o.ID,
			Message:     logger.MsgBusinessValidationFailed,
			Err:         err,
		})
		return err
	}

	err := s.repo.Create(ctx, o)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityCreate,
			EntityType:  "order",
			EntityID:    o.ID,
			Message:     logger.MsgBusinessCreateFailed,
			Err:         err,
		})
		return err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityCreate,
		EntityType:  "order",
		EntityID:    o.ID,
		Message:     logger.MsgBusinessCreated,
		Err:         nil,
	})
	return nil
}

func (s *orderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	if id == "" {
		err := errors.New(logger.MsgBusinessInvalidIdentifier)
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "order",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return nil, err
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "order",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return nil, err
	}
	if order == nil {
		err := errors.New(logger.MsgBusinessNotFound) // fixed: not found, not invalid identifier
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "order",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return nil, err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityGet,
		EntityType:  "order",
		EntityID:    id,
		Message:     logger.MsgBusinessRetrieved,
		Err:         nil,
	})
	return order, nil
}

func (s *orderService) GetByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	orders, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityList,
			EntityType:  "order",
			EntityID:    "",
			Message:     logger.MsgBusinessListFailed,
			Err:         err,
		})
		return nil, err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityList,
		EntityType:  "order",
		EntityID:    "",
		Message:     logger.MsgBusinessListed, // fixed: success message
		Err:         nil,
	})
	return orders, nil
}

// UpdateStatus – implements status change with business validation
func (s *orderService) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	if id == "" {
		err := errors.New(logger.MsgBusinessInvalidIdentifier)
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "order",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return err
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "order",
			EntityID:    id,
			Message:     "failed to fetch order",
			Err:         err,
		})
		return err
	}
	if order == nil {
		err := errors.New(logger.MsgBusinessNotFound)
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "order",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return err
	}

	if err := order.ValidateForUpdateStatus(status); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "order",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "order",
			EntityID:    id,
			Message:     logger.MsgBusinessUpdateFailed,
			Err:         err,
		})
		return err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityUpdate,
		EntityType:  "order",
		EntityID:    id,
		Message:     logger.MsgBusinessUpdated,
		Err:         nil,
	})
	return nil
}

func (s *orderService) List(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.repo.List(ctx)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityList,
			EntityType:  "order",
			EntityID:    "",
			Message:     logger.MsgBusinessListFailed,
			Err:         err,
		})
		return nil, err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityList,
		EntityType:  "order",
		EntityID:    "",
		Message:     logger.MsgBusinessListed,
		Err:         nil,
	})
	return orders, nil
}
