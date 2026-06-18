package postgres

import (
	"context"
	"errors"

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

// Create inserts a new outbox row inside the caller's transaction.
// published_at is deliberately excluded from the INSERT — it must stay NULL
// until the relay worker confirms broker delivery.
func (r *PostgresOutboxRepository) Create(ctx context.Context, tx pgx.Tx, o domain.Outbox) error {
	const q = `
		INSERT INTO outbox (id, event_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at`

	return tx.QueryRow(ctx, q,
		o.ID, o.EventID, o.EventType, o.Payload,
	).Scan(&o.CreatedAt)
}

// List fetches up to limit unpublished rows ordered by creation time.
//
// FOR UPDATE SKIP LOCKED is the critical clause:
//   - FOR UPDATE: locks each row so other relay instances can't pick it up.
//   - SKIP LOCKED: instead of waiting, skips rows already locked by a peer.
//
// Together they allow multiple relay instances to safely share the work
// without publishing the same row twice.
func (r *PostgresOutboxRepository) List(ctx context.Context, limit int) ([]domain.Outbox, error) {
	const q = `
		SELECT id, event_id, event_type, payload, created_at
		FROM   outbox
		WHERE  published_at IS NULL
		ORDER  BY created_at
		LIMIT  $1
		FOR UPDATE SKIP LOCKED`

	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outboxRows []domain.Outbox
	for rows.Next() {
		var o domain.Outbox
		if err := rows.Scan(
			&o.ID, &o.EventID, &o.EventType, &o.Payload, &o.CreatedAt,
		); err != nil {
			return nil, err
		}
		outboxRows = append(outboxRows, o)
	}
	return outboxRows, rows.Err()
}

// MarkPublished sets published_at = NOW() after the broker confirms delivery.
// If the row was deleted between List and MarkPublished (shouldn't happen,
// but defensive), ErrNoRows is swallowed — the event was published either way.
func (r *PostgresOutboxRepository) MarkPublished(ctx context.Context, id string) error {
	const q = `UPDATE outbox SET published_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}
