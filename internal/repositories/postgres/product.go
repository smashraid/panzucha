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

	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	query := `INSERT INTO products (id, name, price, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	result, err := r.pool.Exec(ctx,
		query,
		p.ID, p.Name, p.Price, p.CreatedAt, p.UpdatedAt,
	)

	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBInsert,
		Table:        "products",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom: map[string]any{
			"method": "create",
			"query":  query,
		},
	})
	return err
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	start := time.Now()
	var p domain.Product
	query := `SELECT id, name, price FROM products WHERE id=$1`
	err := r.pool.QueryRow(ctx,
		query,
		id,
	).Scan(&p.ID, &p.Name, &p.Price)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if errors.Is(err, sql.ErrNoRows) {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "products",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          nil,
			Custom: map[string]any{
				"method": "get_by_id",
				"query":  query,
			},
		})
		return nil, nil
	}

	if err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "products",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom: map[string]any{
				"method": "get_by_id",
				"query":  query,
			},
		})
		return nil, err
	}

	rowsAffected = 1
	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "products",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
		Custom: map[string]any{
			"method": "get_by_id",
			"query":  query,
		},
	})
	return &p, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	start := time.Now()
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBDelete,
		Table:        "products",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom: map[string]any{
			"method": "delete",
			"query":  query,
		},
	})

	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New(logger.MsgBusinessNotFound)
	}
	return nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *domain.Product) error {
	start := time.Now()
	now := time.Now().UTC()
	p.UpdatedAt = now
	query := `UPDATE products SET name = $1, price = $2, updated_at = $3 WHERE id = $4`
	result, err := r.pool.Exec(ctx,
		query,
		p.Name, p.Price, p.UpdatedAt, p.ID)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBUpdate,
		Table:        "products",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom: map[string]any{
			"method": "update",
			"query":  query,
		},
	})

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New(logger.MsgBusinessNotFound)
	}

	return nil
}

func (r *PostgresProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	start := time.Now()
	query := `SELECT id, name, price FROM products`
	rows, err := r.pool.Query(ctx, query)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "products",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom: map[string]any{
				"method": "list",
				"query":  query,
			},
		})
		return nil, err
	}
	defer rows.Close()
	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			duration = time.Since(start)
			rowsAffected = int64(len(products))
			r.logger.LogDB(logger.DBLogParams{
				Ctx:          ctx,
				Operation:    logger.DBSelect,
				Table:        "products",
				Duration:     duration,
				RowsAffected: rowsAffected,
				Err:          err,
				Custom: map[string]any{
					"method": "list",
					"query":  query,
				},
			})
			return nil, err
		}
		products = append(products, p)
	}

	duration = time.Since(start)
	rowsAffected = int64(len(products))

	if err := rows.Err(); err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "products",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom: map[string]any{
				"method": "list",
				"query":  query,
			},
		})
		return nil, err
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "products",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
		Custom: map[string]any{
			"method": "list",
			"query":  query,
		},
	})
	return products, nil
}

func (r *PostgresProductRepository) DecrementStock(ctx context.Context, id string, quantity int) error {
	start := time.Now()
	query := `UPDATE products SET stock = stock - $1, updated_at = NOW() WHERE id = $2 AND stock >= $1`
	result, err := r.pool.Exec(ctx, query, quantity, id)
	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = result.RowsAffected()
	}
	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBUpdate,
		Table:        "products",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom:       map[string]any{"method": "decrement_stock", "quantity": quantity},
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}
