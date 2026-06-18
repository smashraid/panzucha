package postgres

import (
	"context"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresInboxRepository struct {
	pool *pgxpool.Pool
}

var _ domain.InboxRepository = (*PostgresInboxRepository)(nil)

func NewPostgresInboxRepository(pool *pgxpool.Pool) *PostgresInboxRepository {
	return &PostgresInboxRepository{pool: pool}
}

func (r *PostgresInboxRepository) Create(ctx context.Context, tx pgx.Tx, i domain.Inbox) error {
	const q = `
        INSERT INTO inbox (event_id, processed_at)
        VALUES ($1, NOW())
        ON CONFLICT (event_id) DO NOTHING`

	tag, err := tx.Exec(ctx, q, i.EventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}
