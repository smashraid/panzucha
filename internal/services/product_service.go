package services

import (
	"context"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/repositories"
)

type ProductService interface {
	Create(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id int) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]domain.Product, error)
}

type productService struct {
	repo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, p *domain.Product) error {
	// business validation
	if p.Name == "" {
		return errors.New("product name cannot be empty")
	}
	if p.Price <= 0 {
		return errors.New("price must be positive")
	}
	return s.repo.Create(ctx, p)
}

func (s *productService) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	if id <= 0 {
		return nil, errors.New("invalid product id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *productService) Update(ctx context.Context, p *domain.Product) error {
	if p.ID <= 0 {
		return errors.New("invalid product id")
	}
	if p.Name == "" {
		return errors.New("product name cannot be empty")
	}
	if p.Price <= 0 {
		return errors.New("price must be positive")
	}
	return s.repo.Update(ctx, p)
}

func (s *productService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid product id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *productService) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.List(ctx)
}
