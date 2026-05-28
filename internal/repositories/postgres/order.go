package repositories

import (
	"context"
	"errors"
	"panzucha/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

var _ domain.OrderRepository = (*PostgresOrderRepository)(nil)

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Fetch product and lock row to prevent concurrent race conditions
	var stock, version int
	err = tx.QueryRow(ctx, "SELECT stock, version FROM products WHERE id = $1 FOR UPDATE", o.ProductID).Scan(&stock, &version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}

	// 2. Validate stock level
	if stock < o.Quantity {
		return domain.ErrInsufficientStock
	}

	// 3. Decrement stock and update version
	_, err = tx.Exec(ctx, "UPDATE products SET stock = stock - $1, version = version + 1, updated_at = NOW() WHERE id = $2", o.Quantity, o.ProductID)
	if err != nil {
		return err
	}

	// 4. Insert order record
	if o.ID == "" {
		o.ID = domain.NewOrderID()
	}

	now := time.Now().UTC()
	o.Audit.CreatedAt = now
	o.Audit.UpdatedAt = now

	query := `INSERT INTO orders (id, user_id, product_id, quantity, total_price, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.Exec(ctx, query,
		o.ID, o.UserID, o.ProductID, o.Quantity, o.TotalPrice, o.Status, o.Audit.CreatedAt, o.Audit.UpdatedAt)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	return err
}

// GetByID retrieves a single order by ID.
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var o domain.Order
	query := `SELECT id, user_id, product_id, quantity, total_price, status, created_at, updated_at
              FROM orders WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status,
		&o.Audit.CreatedAt, &o.Audit.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &o, nil
}

// ListByUser returns all orders for a specific user with pagination.
func (r *PostgresOrderRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Order, error) {
	query := `SELECT id, user_id, product_id, quantity, total_price, status, created_at, updated_at
              FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status,
			&o.Audit.CreatedAt, &o.Audit.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

// UpdateStatus changes the status of an order and updates updated_at.
func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	q := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.pool.Exec(ctx, q, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List returns all orders (admin use).
func (r *PostgresOrderRepository) List(ctx context.Context) ([]domain.Order, error) {
	query := `SELECT id, user_id, product_id, quantity, total_price, status, created_at, updated_at
              FROM orders ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status,
			&o.Audit.CreatedAt, &o.Audit.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
