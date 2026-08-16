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

func (r *PostgresOutboxRepository) Create(ctx context.Context, tx pgx.Tx, o domain.Outbox) error {
	const q = `
		INSERT INTO outbox (id, event_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at`

	return tx.QueryRow(ctx, q,
		o.ID, o.EventID, o.EventType, o.Payload,
	).Scan(&o.CreatedAt)
}

// ListAndLock fetches and locks up to limit rows in ONE round-trip.
// FOR UPDATE SKIP LOCKED still applies to the whole result set — every row
// returned is locked; rows already locked by a concurrent relay instance
// are transparently skipped, never returned, never double-processed.
//
// The lock on every returned row is held until the caller commits or rolls
// back the transaction — which is why the caller must publish and call
// MarkPublishedBatch/MarkFailedBatch BEFORE calling tx.Commit.
func (r *PostgresOutboxRepository) ListAndLock(ctx context.Context, tx pgx.Tx, limit, maxRetries int) ([]domain.Outbox, error) {
	const q = `
		SELECT id, event_id, event_type, payload, retry_count, last_error, created_at
		FROM   outbox
		WHERE  published_at IS NULL
		AND    retry_count < $1
		ORDER  BY created_at
		LIMIT  $2
		FOR UPDATE SKIP LOCKED`

	rows, err := tx.Query(ctx, q, maxRetries, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Outbox
	for rows.Next() {
		var o domain.Outbox
		if err := rows.Scan(
			&o.ID, &o.EventID, &o.EventType, &o.Payload,
			&o.RetryCount, &o.LastError, &o.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

// MarkPublishedBatch updates every row in ids with a single UPDATE ... = ANY($1).
// One round-trip regardless of how many rows succeeded.
func (r *PostgresOutboxRepository) MarkPublishedBatch(ctx context.Context, tx pgx.Tx, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	const q = `UPDATE outbox SET published_at = NOW() WHERE id = ANY($1)`
	_, err := tx.Exec(ctx, q, ids)
	return err
}

// MarkFailedBatch increments retry_count and sets a shared error message
// for every row in ids in a single round-trip.
func (r *PostgresOutboxRepository) MarkFailedBatch(ctx context.Context, tx pgx.Tx, ids []string, errMsg string) error {
	if len(ids) == 0 {
		return nil
	}
	const q = `
		UPDATE outbox
		SET    retry_count = retry_count + 1,
		       last_error  = $1
		WHERE  id = ANY($2)`
	_, err := tx.Exec(ctx, q, errMsg, ids)
	return err
}
