package postgres

import (
	"context"
	"errors"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresIdempotencyKeyRepository struct {
	pool *pgxpool.Pool
}

var _ domain.IdempotencyKeyRepository = (*PostgresIdempotencyKeyRepository)(nil)

func NewPostgresIdempotencyKeyRepository(pool *pgxpool.Pool) *PostgresIdempotencyKeyRepository {
	return &PostgresIdempotencyKeyRepository{pool: pool}
}

// Create inserts a new key with status "processing".
// Uses INSERT ... ON CONFLICT DO NOTHING and checks RowsAffected to detect
// a duplicate — if another request already inserted this key (race condition
// on simultaneous first requests), we return ErrConflict so the handler
// knows to check FindByKey for the in-flight status.
func (r *PostgresIdempotencyKeyRepository) Create(ctx context.Context, key *domain.IdempotencyKey) error {
	const q = `
		INSERT INTO idempotency_keys (key, resource_type, status, created_at, expires_at)
		VALUES ($1, $2, $3, NOW(), $4)
		ON CONFLICT (key) DO NOTHING`

	tag, err := r.pool.Exec(ctx, q,
		key.Key, key.ResourceType, domain.IdempotencyStatusProcessing, key.ExpiresAt,
	)
	if err != nil {
		return err
	}
	// RowsAffected == 0 means the key already exists.
	// The handler must call FindByKey to determine whether it is
	// "processing" (concurrent duplicate) or "completed" (safe replay).
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

// FindByKey retrieves a non-expired idempotency key.
// Expired keys are excluded at the query level — they are logically gone
// even if the cleanup job hasn't deleted them yet.
func (r *PostgresIdempotencyKeyRepository) FindByKey(ctx context.Context, key string) (*domain.IdempotencyKey, error) {
	const q = `
		SELECT key, resource_type, resource_id, response_status, response_body, status, created_at, expires_at
		FROM   idempotency_keys
		WHERE  key = $1
		AND    expires_at > NOW()`

	var k domain.IdempotencyKey
	// resource_id, response_status, response_body are NULL until UpdateToCompleted
	// is called — scan into pointers so pgx handles NULL without panicking.
	var resourceID *string
	var responseStatus *int
	var responseBody []byte

	err := r.pool.QueryRow(ctx, q, key).Scan(
		&k.Key, &k.ResourceType, &resourceID,
		&responseStatus, &responseBody,
		&k.Status, &k.CreatedAt, &k.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Safely dereference nullable fields.
	if resourceID != nil {
		k.ResourceID = *resourceID
	}
	if responseStatus != nil {
		k.ResponseStatus = *responseStatus
	}
	k.ResponseBody = responseBody

	return &k, nil
}

// UpdateToCompleted transitions the key from "processing" → "completed" and
// stores the final response so future duplicate requests receive a replay.
// Called by the service after a successful order creation — inside the same
// transaction so the order row and the key update commit together or not at all.
func (r *PostgresIdempotencyKeyRepository) UpdateToCompleted(
	ctx context.Context,
	tx pgx.Tx,
	key string,
	resourceID string,
	statusCode int,
	responseBody []byte,
) error {
	const q = `
		UPDATE idempotency_keys
		SET    status          = $1,
		       resource_id     = $2,
		       response_status = $3,
		       response_body   = $4
		WHERE  key = $5`

	tag, err := tx.Exec(ctx, q,
		domain.IdempotencyStatusCompleted,
		resourceID, statusCode, responseBody,
		key,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a key when order creation fails so the client can retry
// with the same key. Called outside any transaction — by the time we call
// Delete the business transaction has already rolled back.
func (r *PostgresIdempotencyKeyRepository) Delete(ctx context.Context, key string) error {
	const q = `DELETE FROM idempotency_keys WHERE key = $1`
	_, err := r.pool.Exec(ctx, q, key)
	return err
}
