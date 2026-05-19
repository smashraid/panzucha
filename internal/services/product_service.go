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
		s.logger.LogBusiness(ctx, "product_creation", "product", "", err.Error(), err)
		return err
	}

	err := s.repo.Create(ctx, p)
	if err != nil {
		s.logger.LogBusiness(ctx, "product_creation", "product", p.ID, "failed to create product", err)
		return err
	}

	s.logger.LogBusiness(ctx, "product_creation", "product", p.ID, "product created successfully", nil)
	return nil
}

func (s *productService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		err := errors.New("invalid product id")
		s.logger.LogBusiness(ctx, "product_get", "product", "", err.Error(), err)
		return nil, err
	}

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(ctx, "product_get", "product", id, "database error", err)
		return nil, err
	}
	if product == nil {
		s.logger.LogBusiness(ctx, "product_get", "product", id, "product not found", nil)
		return nil, errors.New("product not found")
	}

	s.logger.LogBusiness(ctx, "product_get", "product", id, "product retrieved", nil)
	return product, nil
}

func (s *productService) Update(ctx context.Context, p *domain.Product) error {
	if p.ID == "" {
		err := errors.New("product id is required")
		s.logger.LogBusiness(ctx, "product_update", "product", "", err.Error(), err)
		return err
	}

	if err := p.ValidateForUpdate(); err != nil {
		s.logger.LogBusiness(ctx, "product_update", "product", p.ID, err.Error(), err)
		return err
	}

	_, err := s.productExists(ctx, p.ID)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, p)
	if err != nil {
		s.logger.LogBusiness(ctx, "product_update", "product", p.ID, "failed to update product", err)
		return err
	}

	s.logger.LogBusiness(ctx, "product_update", "product", p.ID, "product updated successfully", nil)
	return nil
}

func (s *productService) Delete(ctx context.Context, id string) error {
	if id == "" {
		err := errors.New("invalid product id")
		s.logger.LogBusiness(ctx, "product_delete", "product", "", err.Error(), err)
		return err
	}

	_, err := s.productExists(ctx, id)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.LogBusiness(ctx, "product_delete", "product", id, "failed to delete product", err)
		return err
	}

	s.logger.LogBusiness(ctx, "product_delete", "product", id, "product deleted successfully", nil)
	return nil
}

func (s *productService) List(ctx context.Context) ([]domain.Product, error) {
	products, err := s.repo.List(ctx)
	if err != nil {
		s.logger.LogBusiness(ctx, "product_list", "product", "", "failed to list products", err)
		return nil, err
	}

	s.logger.LogBusiness(ctx, "product_list", "product", "", "products listed successfully", nil)
	return products, nil
}

func (s *productService) productExists(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.LogBusiness(ctx, "product_exists", "product", id, "failed to fetch product", err)
		return nil, err
	}
	if product == nil {
		err := errors.New("product not found")
		s.logger.LogBusiness(ctx, "product_exists", "product", id, err.Error(), err)
		return nil, err
	}
	return product, nil
}
