package outbox

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
	RetryCount  int
	LastError   *string
	CreatedAt   time.Time
}

type OutboxRepository interface {
	Create(ctx context.Context, tx pgx.Tx, o Outbox) error

	// ListAndLock fetches up to limit unpublished, retryable rows inside tx,
	// locking ALL of them with FOR UPDATE SKIP LOCKED in a single round-trip.
	// The lock is held until the caller commits or rolls back the tx — so the
	// caller must publish and call MarkPublished/MarkFailed BEFORE committing.
	ListAndLock(ctx context.Context, tx pgx.Tx, limit, maxRetries int) ([]Outbox, error)

	// MarkPublishedBatch sets published_at = NOW() for all given IDs in one
	// statement using = ANY($1) — one round-trip regardless of batch size.
	MarkPublishedBatch(ctx context.Context, tx pgx.Tx, ids []string) error

	// MarkFailedBatch increments retry_count and sets last_error for all
	// given IDs in one statement. errMsg is shared because batch publish
	// failures are usually the same root cause (broker down, etc.); if you
	// need per-row error messages, fall back to individual MarkFailed calls
	// for that subset.
	MarkFailedBatch(ctx context.Context, tx pgx.Tx, ids []string, errMsg string) error
}
