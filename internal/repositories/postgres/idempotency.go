package repositories

import (
	"context"
	"database/sql"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresIdempotencyKeyRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewPostgresIdempotencyKeyRepository(pool *pgxpool.Pool, log *logger.Logger) *PostgresIdempotencyKeyRepository {
	return &PostgresIdempotencyKeyRepository{pool: pool, logger: log}
}

func (r *PostgresIdempotencyKeyRepository) Create(ctx context.Context, key *domain.IdempotencyKey) error {
	query := `
        INSERT INTO idempotency_keys (key, resource_type, resource_id, response_status, response_body, status, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.pool.Exec(ctx, query,
		key.Key, key.ResourceType, key.ResourceID, key.ResponseStatus, key.ResponseBody, key.Status,
		key.CreatedAt, key.ExpiresAt,
	)
	return err
}

func (r *PostgresIdempotencyKeyRepository) UpdateToCompleted(ctx context.Context, key string, resourceID string, statusCode int, responseBody []byte) error {
	query := `
        UPDATE idempotency_keys
        SET resource_id = $2, response_status = $3, response_body = $4, status = 'completed'
        WHERE key = $1 AND status = 'processing'
    `
	_, err := r.pool.Exec(ctx, query, key, resourceID, statusCode, responseBody)
	return err
}

func (r *PostgresIdempotencyKeyRepository) FindByKey(ctx context.Context, keyStr string) (*domain.IdempotencyKey, error) {
	query := `
        SELECT key, resource_type, resource_id, response_status, response_body, created_at, expires_at
        FROM idempotency_keys
        WHERE key = $1 AND expires_at > NOW()
    `
	var key domain.IdempotencyKey
	var responseBody []byte
	err := r.pool.QueryRow(ctx, query, keyStr).Scan(
		&key.Key, &key.ResourceType, &key.ResourceID, &key.ResponseStatus, &responseBody,
		&key.CreatedAt, &key.ExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	key.ResponseBody = responseBody
	return &key, nil
}

func (r *PostgresIdempotencyKeyRepository) UpdateWithResponse(ctx context.Context, keyStr string, resourceID string, statusCode int, responseBody []byte) error {
	query := `
        UPDATE idempotency_keys
        SET resource_id = $2, response_status = $3, response_body = $4
        WHERE key = $1
    `
	_, err := r.pool.Exec(ctx, query, keyStr, resourceID, statusCode, responseBody)
	return err
}

func (r *PostgresIdempotencyKeyRepository) Delete(ctx context.Context, keyStr string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM idempotency_keys WHERE key = $1", keyStr)
	return err
}
