package domain

import (
	"context"
	"time"
)

type IdempotencyStatus string

const (
	IdempotencyStatusProcessing IdempotencyStatus = "processing"
	IdempotencyStatusCompleted  IdempotencyStatus = "completed"
)

type IdempotencyKey struct {
	Key            string
	ResourceType   string
	ResourceID     string
	ResponseStatus int
	ResponseBody   []byte
	Status         IdempotencyStatus
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

type IdempotencyKeyRepository interface {
	// Insert a new key with status "processing"
	Create(ctx context.Context, key *IdempotencyKey) error
	// Find by key (non-expired)
	FindByKey(ctx context.Context, key string) (*IdempotencyKey, error)
	// Update to "completed" with final response
	UpdateToCompleted(ctx context.Context, key string, resourceID string, statusCode int, responseBody []byte) error
	// Delete key if processing fails
	Delete(ctx context.Context, key string) error
}
