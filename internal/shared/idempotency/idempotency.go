package idempotency

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type IdempotencyStatus string

const (
	IdempotencyStatusProcessing IdempotencyStatus = "processing"
	IdempotencyStatusCompleted  IdempotencyStatus = "completed"
)

type IdempotencyKey struct {
	Key            string
	ResourceType   string
	ResourceID     string // filled after completion
	ResponseStatus int    // HTTP status of the original response
	ResponseBody   []byte // full JSON body to replay
	Status         IdempotencyStatus
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

type IdempotencyKeyRepository interface {
	// Create inserts a new key with status "processing".
	// Returns ErrConflict if the key already exists.
	Create(ctx context.Context, key *IdempotencyKey) error

	// FindByKey returns the key if it exists and has not expired.
	// Returns ErrNotFound if missing or expired.
	FindByKey(ctx context.Context, key string) (*IdempotencyKey, error)

	// UpdateToCompleted transitions the key to "completed" inside the provided
	// transaction, storing the final HTTP response for future replays.
	UpdateToCompleted(ctx context.Context, tx pgx.Tx, key string, resourceID string, statusCode int, responseBody []byte) error

	// Delete removes the key so the client can retry with the same key after
	// a failed request. Called after the business transaction has rolled back.
	Delete(ctx context.Context, key string) error
}
