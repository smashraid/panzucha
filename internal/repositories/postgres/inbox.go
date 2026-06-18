package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"panzucha/internal/domain"
)

var _ domain.InboxRepository = (*PostgresInboxRepository)(nil)

type PostgresInboxRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresInboxRepository(pool *pgxpool.Pool) *PostgresInboxRepository {
	return &PostgresInboxRepository{pool: pool}
}

// Create inserts an event_id inside the caller's transaction.
//
// ON CONFLICT DO NOTHING means a duplicate event_id is silently ignored
// at the DB level. RowsAffected() == 0 is the signal the consumer uses
// to detect a duplicate and skip processing.
//
// Returns domain.ErrConflict on duplicate so the consumer can handle it
// with a simple errors.Is check — no raw RowsAffected logic in the consumer.
//
// The tx parameter is mandatory — the inbox insert and the business logic
// it protects must commit together. If business logic fails and tx rolls back,
// the inbox row is also rolled back, allowing a clean retry on redelivery.
func (r *PostgresInboxRepository) Create(ctx context.Context, tx pgx.Tx, eventID string) error {
	const q = `
		INSERT INTO inbox (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT (event_id) DO NOTHING`

	tag, err := tx.Exec(ctx, q, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict // already processed — consumer should ack and skip
	}
	return nil
}
