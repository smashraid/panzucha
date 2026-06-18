package postgres

import (
	"context"
	"errors"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgrestOutboxRepository struct {
	pool *pgxpool.Pool
}

var _ domain.OutboxRepository = (*PostgrestOutboxRepository)(nil)

func NewPostgresOutboxRepository(pool *pgxpool.Pool) *PostgrestOutboxRepository {
	return &PostgrestOutboxRepository{pool: pool}
}

func (r *PostgrestOutboxRepository) GetByID(ctx context.Context, id string) (*domain.Outbox, error) {
	const q = `
	SELECT id, event_id, event_type, payload, published_at, created_at
	FROM   outbox
	WHERE id = $1
	`
	var o domain.Outbox
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&o.ID, &o.EventID, &o.EventType, &o.Payload, &o.PublishedAt, &o.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *PostgrestOutboxRepository) Create(ctx context.Context, tx pgx.Tx, o domain.Outbox) error {
	const q = `
		INSERT INTO outbox
			(id, event_id, event_type, payload, created_at)
		VALUES
			($1, $2, $3, $4, NOW())		
		RETURNING published_at, created_at
	`
	return tx.QueryRow(ctx, q,
		o.ID, o.EventID, o.EventType, o.Payload,
	).Scan(&o.CreatedAt)
}
