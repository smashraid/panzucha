package postgres

import (
	"context"
	"encoding/json"
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

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	const q = `
		SELECT id, user_id, items, status, total_amount,
		       created_at, created_by, updated_at, updated_by
		FROM   orders
		WHERE  id = $1`

	var o domain.Order
	var itemsJSON []byte

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&o.ID, &o.UserID, &itemsJSON, &o.Status, &o.TotalAmount,
		&o.Audit.CreatedAt, &o.Audit.CreatedBy,
		&o.Audit.UpdatedAt, &o.Audit.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *PostgresOrderRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Order, error) {
	const q = `
		SELECT id, user_id, items, status, total_amount,
		       created_at, created_by, updated_at, updated_by
		FROM   orders
		WHERE  user_id = $1
		ORDER  BY created_at DESC
		LIMIT  $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		var itemsJSON []byte

		if err := rows.Scan(
			&o.ID, &o.UserID, &itemsJSON, &o.Status, &o.TotalAmount,
			&o.Audit.CreatedAt, &o.Audit.CreatedBy,
			&o.Audit.UpdatedAt, &o.Audit.UpdatedBy,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// Create persists a new order inside the provided transaction.
// It receives a pgx.Tx instead of using the pool directly because order
// creation is always part of a larger transaction that also decrements stock.
// The caller (service layer) owns the transaction lifecycle — begin, commit,
// rollback. The repository never calls Begin() or Commit() itself.
func (r *PostgresOrderRepository) Create(ctx context.Context, tx pgx.Tx, o *domain.Order) error {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO orders (id, user_id, items, status, total_amount, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6, NOW(), $6)
		RETURNING created_at, updated_at`

	return tx.QueryRow(ctx, q,
		o.ID, o.UserID, itemsJSON, o.Status, o.TotalAmount, o.Audit.CreatedBy,
	).Scan(&o.Audit.CreatedAt, &o.Audit.UpdatedAt)
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	const q = `
		UPDATE orders
		SET    status     = $1,
		       updated_at = NOW()
		WHERE  id = $2
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.pool.QueryRow(ctx, q, status, id).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}
