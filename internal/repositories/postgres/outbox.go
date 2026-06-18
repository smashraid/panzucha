package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"panzucha/internal/domain"
)

var _ domain.OutboxRepository = (*PostgresOutboxRepository)(nil)

type PostgresOutboxRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOutboxRepository(pool *pgxpool.Pool) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{pool: pool}
}

// Create inserts an outbox row inside the caller's transaction.
// published_at is excluded — stays NULL until the relay marks it.
func (r *PostgresOutboxRepository) Create(ctx context.Context, tx pgx.Tx, o domain.Outbox) error {
	const q = `
		INSERT INTO outbox (id, event_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at`

	return tx.QueryRow(ctx, q,
		o.ID, o.EventID, o.EventType, o.Payload,
	).Scan(&o.CreatedAt)
}

// ListAndLock fetches up to limit unpublished rows inside tx.
// FOR UPDATE: holds a row-level lock until tx commits/rolls back.
// SKIP LOCKED: concurrent relay instances skip locked rows rather than
// waiting — each instance works on a disjoint set of rows.
func (r *PostgresOutboxRepository) ListAndLock(ctx context.Context, tx pgx.Tx, limit int) ([]domain.Outbox, error) {
	const q = `
		SELECT id, event_id, event_type, payload, created_at
		FROM   outbox
		WHERE  published_at IS NULL
		ORDER  BY created_at
		LIMIT  $1
		FOR UPDATE SKIP LOCKED`

	rows, err := tx.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Outbox
	for rows.Next() {
		var o domain.Outbox
		if err := rows.Scan(
			&o.ID, &o.EventID, &o.EventType, &o.Payload, &o.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

// MarkPublished sets published_at inside the same tx that locked the row.
// The lock is released when the caller commits — after broker delivery is confirmed.
func (r *PostgresOutboxRepository) MarkPublished(ctx context.Context, tx pgx.Tx, id string) error {
	const q = `UPDATE outbox SET published_at = NOW() WHERE id = $1`
	_, err := tx.Exec(ctx, q, id)
	return err
}
