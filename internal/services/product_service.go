package services

import (
	"context"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
)

type ProductService interface {
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	List(ctx context.Context, limit, offset int) ([]domain.Product, error)
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id string) error
	DecrementStock(ctx context.Context, id string, qty, version int) error
}

type productService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, p *domain.Product) error {
	if p.Price <= 0 {
		return domain.ErrInvalidInput
	}
	return s.repo.Create(ctx, p)
}

func (s *productService) Update(ctx context.Context, p *domain.Product) error {
	if p.Price <= 0 {
		return domain.ErrInvalidInput
	}
	return s.repo.Update(ctx, p)
}

func (s *productService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) List(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *productService) productExists(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		err := errors.New(logger.MsgBusinessNotFound)
		return nil, err
	}
	return product, nil
}

func (s *productService) DecrementStock(ctx context.Context, id string, qty, version int) error {
	return s.repo.DecrementStock(ctx, id, qty, version)
}
