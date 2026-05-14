package services

import (
	"context"
	"errors"
	"panzucha/internal/domain"
)

type ProductService interface {
	Create(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.Product, error)
}

type productService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, p *domain.Product) error {
	if err := p.ValidateForCreate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, p)
}

func (s *productService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if id == "" {
		return nil, errors.New("invalid product id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *productService) Update(ctx context.Context, p *domain.Product) error {
	if err := p.ValidateForUpdate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, p)
}

func (s *productService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("invalid product id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *productService) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}
