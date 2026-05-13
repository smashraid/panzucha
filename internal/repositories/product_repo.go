package repositories

import (
	"context"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id int) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]domain.Product, error)
}

type postgresProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &postgresProductRepo{pool: pool}
}

func (r *postgresProductRepo) Create(ctx context.Context, p *domain.Product) error {
	return r.pool.QueryRow(ctx,
		"INSERT INTO products (name, price) VALUES ($1, $2) RETURNING id",
		p.Name, p.Price,
	).Scan(&p.ID)
}

func (r *postgresProductRepo) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	var p domain.Product
	err := r.pool.QueryRow(ctx,
		"SELECT id, name, price FROM products WHERE id=$1",
		id,
	).Scan(&p.ID, &p.Name, &p.Price)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresProductRepo) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	return err
}

func (r *postgresProductRepo) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE products SET name = $1, price = $2 WHERE id = $3",
		p.Name, p.Price, p.ID)
	return err
}

func (r *postgresProductRepo) List(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, name, price FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
