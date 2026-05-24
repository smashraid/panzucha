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

type PostgresOrderRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

var _ domain.OrderRepository = (*PostgresOrderRepository)(nil)

func NewPostgresOrderRepository(pool *pgxpool.Pool, log *logger.Logger) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool, logger: log}
}

// Create inserts a new order.
func (r *PostgresOrderRepository) Create(ctx context.Context, o *domain.Order) error {
	start := time.Now()
	if o.ID == "" {
		o.ID = domain.NewOrderID()
	}

	now := time.Now().UTC()
	o.CreatedAt = now
	o.UpdatedAt = now

	query := `INSERT INTO orders (id, user_id, product_id, quantity, total_price, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		o.ID, o.UserID, o.ProductID, o.Quantity, o.TotalPrice, o.Status, o.CreatedAt, o.UpdatedAt)

	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = 1
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBInsert,
		Table:        "orders",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom:       map[string]any{"method": "create"},
	})
	return err
}

// GetByID retrieves a single order by ID.
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	start := time.Now()
	var o domain.Order
	query := `SELECT id, user_id, product_id, quantity, total_price, status, created_at, updated_at
              FROM orders WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status,
		&o.CreatedAt, &o.UpdatedAt,
	)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if errors.Is(err, sql.ErrNoRows) {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "orders",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          nil,
			Custom:       map[string]any{"method": "get_by_id"},
		})
		return nil, nil
	}
	if err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "orders",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom:       map[string]any{"method": "get_by_id"},
		})
		return nil, err
	}

	rowsAffected = 1
	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "orders",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
		Custom:       map[string]any{"method": "get_by_id"},
	})
	return &o, nil
}

// GetByUserID returns all orders for a specific user.
func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	start := time.Now()
	query := `SELECT id, user_id, product_id, quantity, total_price, status, created_at, updated_at
              FROM orders WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "orders",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom:       map[string]any{"method": "get_by_user_id"},
		})
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			duration = time.Since(start)
			rowsAffected = int64(len(orders))
			r.logger.LogDB(logger.DBLogParams{
				Ctx:          ctx,
				Operation:    logger.DBSelect,
				Table:        "orders",
				Duration:     duration,
				RowsAffected: rowsAffected,
				Err:          err,
				Custom:       map[string]any{"method": "get_by_user_id"},
			})
			return nil, err
		}
		orders = append(orders, o)
	}

	duration = time.Since(start)
	rowsAffected = int64(len(orders))

	if err := rows.Err(); err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "orders",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom:       map[string]any{"method": "get_by_user_id"},
		})
		return nil, err
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "orders",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
		Custom:       map[string]any{"method": "get_by_user_id"},
	})
	return orders, nil
}

// UpdateStatus changes the status of an order and updates updated_at.
func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	start := time.Now()
	now := time.Now().UTC()
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.pool.Exec(ctx, query, status, now, id)

	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBUpdate,
		Table:        "orders",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom:       map[string]any{"method": "update_status", "new_status": status},
	})

	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List returns all orders (admin use).
func (r *PostgresOrderRepository) List(ctx context.Context) ([]domain.Order, error) {
	start := time.Now()
	query := `SELECT id, user_id, product_id, quantity, total_price, status, created_at, updated_at
              FROM orders`
	rows, err := r.pool.Query(ctx, query)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "orders",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom:       map[string]any{"method": "list"},
		})
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			duration = time.Since(start)
			rowsAffected = int64(len(orders))
			r.logger.LogDB(logger.DBLogParams{
				Ctx:          ctx,
				Operation:    logger.DBSelect,
				Table:        "orders",
				Duration:     duration,
				RowsAffected: rowsAffected,
				Err:          err,
				Custom:       map[string]any{"method": "list"},
			})
			return nil, err
		}
		orders = append(orders, o)
	}

	duration = time.Since(start)
	rowsAffected = int64(len(orders))

	if err := rows.Err(); err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "orders",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom:       map[string]any{"method": "list"},
		})
		return nil, err
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "orders",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
		Custom:       map[string]any{"method": "list"},
	})
	return orders, nil
}
