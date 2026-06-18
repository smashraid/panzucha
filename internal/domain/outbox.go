package domain

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Outbox struct {
	ID          string
	EventID     string
	EventType   string
	Payload     []byte
	PublishedAt *time.Time
	CreatedAt   time.Time
}

type OutboxRepository interface {
	// Create writes an outbox row inside the caller's transaction.
	// Atomic with the business operation that produced the event.
	Create(ctx context.Context, tx pgx.Tx, o Outbox) error

	// ListAndLock fetches up to limit unpublished rows inside the provided
	// transaction using FOR UPDATE SKIP LOCKED. The lock is held until the
	// transaction commits or rolls back — preventing concurrent relay
	// instances from picking the same rows.
	ListAndLock(ctx context.Context, tx pgx.Tx, limit int) ([]Outbox, error)

	// MarkPublished sets published_at = NOW() inside the same transaction
	// that locked the row, ensuring the lock covers the full publish cycle.
	MarkPublished(ctx context.Context, tx pgx.Tx, id string) error
}
