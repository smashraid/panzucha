package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"panzucha/internal/domain"
)

var _ domain.OrderRepository = (*PostgresOrderRepository)(nil)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

// GetByID fetches a single order with all its items in one round-trip.
//
// json_agg builds the items array server-side. The LEFT JOIN ensures the
// query still returns the order row when it has no items yet (e.g. an order
// that was just inserted but items failed — edge case, but safe).
//
// NULL check on json_agg: if there are no items, json_agg returns NULL,
// not an empty array. We coalesce to '[]' so json.Unmarshal gives us
// an empty slice, not an error.
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	const q = `
		SELECT
			o.id,
			o.user_id,
			o.status,
			o.total_amount,
			o.created_at,
			o.created_by,
			o.updated_at,
			o.updated_by,
			COALESCE(
				json_agg(
					json_build_object(
						'product_id', oi.product_id,
						'quantity',   oi.quantity,
						'unit_price', oi.unit_price
					) ORDER BY oi.created_at
				) FILTER (WHERE oi.id IS NOT NULL),
				'[]'
			) AS items
		FROM  orders o
		LEFT  JOIN order_items oi ON oi.order_id = o.id
		WHERE o.id = $1
		GROUP BY o.id`

	return scanOrder(r.pool.QueryRow(ctx, q, id))
}

// ListByUser fetches all orders for a user with their items, paginated.
// Same json_agg pattern — one round-trip for all orders and their items.
func (r *PostgresOrderRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Order, error) {
	const q = `
		SELECT
			o.id,
			o.user_id,
			o.status,
			o.total_amount,
			o.created_at,
			o.created_by,
			o.updated_at,
			o.updated_by,
			COALESCE(
				json_agg(
					json_build_object(
						'product_id', oi.product_id,
						'quantity',   oi.quantity,
						'unit_price', oi.unit_price
					) ORDER BY oi.created_at
				) FILTER (WHERE oi.id IS NOT NULL),
				'[]'
			) AS items
		FROM  orders o
		LEFT  JOIN order_items oi ON oi.order_id = o.id
		WHERE o.user_id = $1
		GROUP BY o.id
		ORDER BY o.created_at DESC
		LIMIT  $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}
	return orders, rows.Err()
}

// Create inserts the order header and all line items inside the provided
// transaction. The caller (order service) owns the transaction lifecycle.
func (r *PostgresOrderRepository) Create(ctx context.Context, tx pgx.Tx, o *domain.Order) error {
	if err := createOrder(ctx, tx, o); err != nil {
		return err
	}

	for i := range o.Items {
		if _, err := createOrderItem(ctx, tx, o.ID, o.Items[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	const q = `
		UPDATE orders
		SET    status     = $1,
		       updated_at = NOW()
		WHERE  id = $2`

	tag, err := r.pool.Exec(ctx, q, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ── Private helpers ───────────────────────────────────────────────────────────

func createOrder(ctx context.Context, tx pgx.Tx, o *domain.Order) error {
	const q = `
		INSERT INTO orders (id, user_id, status, total_amount, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), $5)
		RETURNING created_at, updated_at`

	return tx.QueryRow(ctx, q,
		o.ID, o.UserID, o.Status, o.TotalAmount, o.Audit.CreatedBy,
	).Scan(&o.Audit.CreatedAt, &o.Audit.UpdatedAt)
}

func createOrderItem(ctx context.Context, tx pgx.Tx, orderID string, oi domain.OrderItem) (pgconn.CommandTag, error) {
	const q = `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`

	return tx.Exec(ctx, q,
		uuid.NewString(), orderID, oi.ProductID, oi.Quantity, oi.UnitPrice,
	)
}

// scanOrder reads one order row (including the json_agg items column) from
// any type that implements the Scan method — works for both pgx.Row (single)
// and pgx.Rows (loop). This avoids duplicating the scan logic.
func scanOrder(row interface {
	Scan(dest ...any) error
}) (*domain.Order, error) {
	var o domain.Order
	var itemsJSON []byte

	err := row.Scan(
		&o.ID, &o.UserID, &o.Status, &o.TotalAmount,
		&o.Audit.CreatedAt, &o.Audit.CreatedBy,
		&o.Audit.UpdatedAt, &o.Audit.UpdatedBy,
		&itemsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if err := unmarshalItems(itemsJSON, &o.Items); err != nil {
		return nil, err
	}
	return &o, nil
}

// unmarshalItems decodes the json_agg result into a []domain.OrderItem.
// Uses an intermediate struct because the JSON keys from json_build_object
// must match exactly — we control them in the query above.
func unmarshalItems(data []byte, items *[]domain.OrderItem) error {
	if len(data) == 0 {
		*items = []domain.OrderItem{}
		return nil
	}

	type itemRow struct {
		ProductID string  `json:"product_id"`
		Quantity  int     `json:"quantity"`
		UnitPrice float64 `json:"unit_price"`
	}

	var rows []itemRow
	if err := unmarshalJSON(data, &rows); err != nil {
		return err
	}

	result := make([]domain.OrderItem, len(rows))
	for i, r := range rows {
		result[i] = domain.OrderItem{
			ProductID: r.ProductID,
			Quantity:  r.Quantity,
			UnitPrice: r.UnitPrice,
		}
	}
	*items = result
	return nil
}
