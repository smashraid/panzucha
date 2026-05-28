package repositories

import (
	"context"
	"errors"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

var _ domain.ProductRepository = (*PostgresProductRepository)(nil)

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	const q = `
		SELECT id, name, description, price, stock, version,
		       created_at, created_by, updated_at, updated_by
		FROM   products
		WHERE  id = $1`

	var p domain.Product
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Name, &p.Description,
		&p.Price, &p.Stock, &p.Version,
		&p.Audit.CreatedAt, &p.Audit.CreatedBy,
		&p.Audit.UpdatedAt, &p.Audit.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, limit, offset int) ([]domain.Product, error) {
	const q = `
		SELECT id, name, description, price, stock, version,
		       created_at, created_by, updated_at, updated_by
		FROM   products
		ORDER  BY created_at DESC
		LIMIT  $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description,
			&p.Price, &p.Stock, &p.Version,
			&p.Audit.CreatedAt, &p.Audit.CreatedBy,
			&p.Audit.UpdatedAt, &p.Audit.UpdatedBy,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *domain.Product) error {
	const q = `
		INSERT INTO products
			(id, name, description, price, stock, version, created_at, created_by, updated_at, updated_by)
		VALUES
			($1, $2, $3, $4, $5, 1, NOW(), $6, NOW(), $6)
		RETURNING version, created_at, updated_at`

	return r.pool.QueryRow(ctx, q,
		p.ID, p.Name, p.Description, p.Price, p.Stock,
		p.Audit.CreatedBy,
	).Scan(&p.Version, &p.Audit.CreatedAt, &p.Audit.UpdatedAt)
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *domain.Product) error {
	const q = `
		UPDATE products
		SET    name        = $1,
		       description = $2,
		       price       = $3,
		       stock       = $4,
		       updated_by  = $5,
		       updated_at  = NOW(),
		       version     = version + 1
		WHERE  id = $6
		RETURNING version, updated_at`

	err := r.pool.QueryRow(ctx, q,
		p.Name, p.Description, p.Price, p.Stock,
		p.Audit.UpdatedBy, p.ID,
	).Scan(&p.Version, &p.Audit.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM products WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostgresProductRepository) DecrementStock(ctx context.Context, id string, qty int, version int) error {
	const q = `
		UPDATE products
		SET    stock   = stock - $1,
		       version = version + 1,
		       updated_at = NOW()
		WHERE  id = $2 AND version = $3 AND stock >= $1`

	tag, err := r.pool.Exec(ctx, q, qty, id, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	// Distinguish the three failure causes with a single follow-up SELECT.
	var current struct {
		Stock   int
		Version int
	}
	const check = `SELECT stock, version FROM products WHERE id = $1`
	err = r.pool.QueryRow(ctx, check, id).Scan(&current.Stock, &current.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if current.Version != version {
		return domain.ErrVersionConflict
	}
	return domain.ErrInsufficientStock
}
