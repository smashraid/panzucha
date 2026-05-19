package repositories

import (
	"context"
	"database/sql"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

var _ domain.ProductRepository = (*PostgresProductRepository)(nil)

func NewPostgresProductRepository(pool *pgxpool.Pool, log *logger.Logger) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool, logger: log}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *domain.Product) error {
	start := time.Now()
	if p.ID == "" {
		p.ID = domain.NewProductID()
	}

	result, err := r.pool.Exec(ctx,
		"INSERT INTO products (id, name, price) VALUES ($1, $2, $3)",
		p.ID, p.Name, p.Price,
	)

	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(ctx, "db_insert", "products", duration, rowsAffected, err)
	return err
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	start := time.Now()
	var p domain.Product
	err := r.pool.QueryRow(ctx,
		"SELECT id, name, price FROM products WHERE id=$1",
		id,
	).Scan(&p.ID, &p.Name, &p.Price)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if errors.Is(err, sql.ErrNoRows) {
		r.logger.LogDB(ctx, "db_select", "products", duration, rowsAffected, nil)
		return nil, nil
	}

	if err != nil {
		r.logger.LogDB(ctx, "db_select", "products", duration, rowsAffected, err)
		return nil, err
	}

	rowsAffected = 1
	r.logger.LogDB(ctx, "db_select", "products", duration, rowsAffected, nil)
	return &p, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	start := time.Now()
	result, err := r.pool.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(ctx, "db_delete", "products", duration, rowsAffected, err)
	return err
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *domain.Product) error {
	start := time.Now()
	result, err := r.pool.Exec(ctx,
		"UPDATE products SET name = $1, price = $2 WHERE id = $3",
		p.Name, p.Price, p.ID)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(ctx, "db_update", "products", duration, rowsAffected, err)
	return err
}

func (r *PostgresProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	start := time.Now()
	rows, err := r.pool.Query(ctx, "SELECT id, name, price FROM products")
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err != nil {
		r.logger.LogDB(ctx, "db_select", "products", duration, rowsAffected, err)
		return nil, err
	}
	defer rows.Close()
	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			duration = time.Since(start)
			rowsAffected = int64(len(products))
			r.logger.LogDB(ctx, "db_select", "products", duration, rowsAffected, err)
			return nil, err
		}
		products = append(products, p)
	}

	duration = time.Since(start)
	rowsAffected = int64(len(products))

	if err := rows.Err(); err != nil {
		r.logger.LogDB(ctx, "db_select", "products", duration, rowsAffected, err)
		return nil, err
	}

	r.logger.LogDB(ctx, "db_select", "products", duration, rowsAffected, nil)
	return products, nil
}
