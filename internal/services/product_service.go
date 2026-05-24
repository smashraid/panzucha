package services

import (
	"context"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
)

type ProductService interface {
	Create(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Product, error)
}

type productService struct {
	repo   domain.ProductRepository
	logger *logger.Logger
}

func NewProductService(repo domain.ProductRepository, log *logger.Logger) ProductService {
	return &productService{repo: repo, logger: log}
}

func (s *productService) Create(ctx context.Context, p *domain.Product) error {
	if err := p.ValidateForCreate(); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityCreate,
			EntityType:  "product",
			EntityID:    p.ID,
			Message:     logger.MsgBusinessValidationFailed,
			Err:         err,
		})
		return err
	}

	err := s.repo.Create(ctx, p)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityCreate,
			EntityType:  "product",
			EntityID:    p.ID,
			Message:     logger.MsgBusinessCreateFailed,
			Err:         err,
		})
		return err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityCreate,
		EntityType:  "product",
		EntityID:    p.ID,
		Message:     logger.MsgBusinessCreated,
		Err:         nil,
	})
	return nil
}

func (s *productService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		err := errors.New(logger.MsgBusinessInvalidIdentifier)
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "product",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return nil, err
	}

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "product",
			EntityID:    id,
			Message:     logger.MsgBusinessDatabaseError,
			Err:         err,
		})
		return nil, err
	}
	if product == nil {
		err := errors.New(logger.MsgBusinessNotFound)
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityGet,
			EntityType:  "product",
			EntityID:    id,
			Message:     err.Error(),
			Err:         nil,
		})
		return nil, err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityGet,
		EntityType:  "product",
		EntityID:    id,
		Message:     logger.MsgBusinessRetrieved,
		Err:         nil,
	})
	return product, nil
}

func (s *productService) Update(ctx context.Context, p *domain.Product) error {
	if p.ID == "" {
		err := errors.New("product id is required")
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "product",
			EntityID:    p.ID,
			Message:     err.Error(),
			Err:         err,
		})
		return err
	}

	if err := p.ValidateForUpdate(); err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "product",
			EntityID:    p.ID,
			Message:     err.Error(),
			Err:         err,
		})
		return err
	}

	_, err := s.productExists(ctx, p.ID)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, p)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityUpdate,
			EntityType:  "product",
			EntityID:    p.ID,
			Message:     logger.MsgBusinessUpdateFailed,
			Err:         err,
		})
		return err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityUpdate,
		EntityType:  "product",
		EntityID:    p.ID,
		Message:     logger.MsgBusinessUpdated,
		Err:         nil,
	})
	return nil
}

func (s *productService) Delete(ctx context.Context, id string) error {
	if id == "" {
		err := errors.New("invalid product id")
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityDelete,
			EntityType:  "product",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return err
	}

	_, err := s.productExists(ctx, id)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityDelete,
			EntityType:  "product",
			EntityID:    id,
			Message:     logger.MsgBusinessDeleteFailed,
			Err:         err,
		})
		return err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityDelete,
		EntityType:  "product",
		EntityID:    id,
		Message:     logger.MsgBusinessDeleted,
		Err:         nil,
	})
	return nil
}

func (s *productService) List(ctx context.Context) ([]domain.Product, error) {
	products, err := s.repo.List(ctx)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessEntityList,
			EntityType:  "product",
			EntityID:    "",
			Message:     logger.MsgBusinessListFailed,
			Err:         err,
		})
		return nil, err
	}

	s.logger.LogBusiness(logger.BusinessLogParams{
		Ctx:         ctx,
		SubCategory: logger.BusinessEntityList,
		EntityType:  "product",
		EntityID:    "",
		Message:     logger.MsgBusinessListed,
		Err:         nil,
	})
	return products, nil
}

func (s *productService) productExists(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessExistEntity,
			EntityType:  "product",
			EntityID:    id,
			Message:     logger.MsgBusinessGetFailed,
			Err:         err,
		})
		return nil, err
	}
	if product == nil {
		err := errors.New(logger.MsgBusinessNotFound)
		s.logger.LogBusiness(logger.BusinessLogParams{
			Ctx:         ctx,
			SubCategory: logger.BusinessExistEntity,
			EntityType:  "product",
			EntityID:    id,
			Message:     err.Error(),
			Err:         err,
		})
		return nil, err
	}
	return product, nil
}
