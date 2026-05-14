package repositories

import (
	"context"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

var _ domain.ProductRepository = (*PostgresProductRepository)(nil)

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *domain.Product) error {
	return r.pool.QueryRow(ctx,
		"INSERT INTO products (name, price) VALUES ($1, $2) RETURNING id",
		p.Name, p.Price,
	).Scan(&p.ID)
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
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

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	return err
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE products SET name = $1, price = $2 WHERE id = $3",
		p.Name, p.Price, p.ID)
	return err
}

func (r *PostgresProductRepository) List(ctx context.Context) ([]domain.Product, error) {
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
