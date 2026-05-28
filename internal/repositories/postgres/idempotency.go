package repositories

import (
	"context"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresIdempotencyKeyRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

var _ domain.IdempotencyKeyRepository = (*PostgresIdempotencyKeyRepository)(nil)

func NewPostgresIdempotencyKeyRepository(pool *pgxpool.Pool, log *logger.Logger) *PostgresIdempotencyKeyRepository {
	return &PostgresIdempotencyKeyRepository{pool: pool, logger: log}
}

func (r *PostgresIdempotencyKeyRepository) Create(ctx context.Context, key *domain.IdempotencyKey) error {
	start := time.Now()
	query := `
        INSERT INTO idempotency_keys (key, resource_type, resource_id, response_status, response_body, status, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.pool.Exec(ctx, query,
		key.Key, key.ResourceType, key.ResourceID, key.ResponseStatus, key.ResponseBody, key.Status,
		key.CreatedAt, key.ExpiresAt,
	)
	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = 1
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBInsert,
		Table:        "idempotency_keys",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom:       map[string]any{"method": "create", "key": key.Key},
	})
	return err
}

func (r *PostgresIdempotencyKeyRepository) UpdateToCompleted(ctx context.Context, key string, resourceID string, statusCode int, responseBody []byte) error {
	start := time.Now()
	query := `
        UPDATE idempotency_keys
        SET resource_id = $2, response_status = $3, response_body = $4, status = 'completed'
        WHERE key = $1 AND status = 'processing'
    `
	tag, err := r.pool.Exec(ctx, query, key, resourceID, statusCode, responseBody)
	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = tag.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBUpdate,
		Table:        "idempotency_keys",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom:       map[string]any{"method": "update_to_completed", "key": key},
	})
	return err
}

func (r *PostgresIdempotencyKeyRepository) FindByKey(ctx context.Context, keyStr string) (*domain.IdempotencyKey, error) {
	start := time.Now()
	query := `
        SELECT key, resource_type, resource_id, response_status, response_body, status, created_at, expires_at
        FROM idempotency_keys
        WHERE key = $1 AND expires_at > NOW()
    `
	var key domain.IdempotencyKey
	var responseBody []byte
	err := r.pool.QueryRow(ctx, query, keyStr).Scan(
		&key.Key, &key.ResourceType, &key.ResourceID, &key.ResponseStatus, &responseBody, &key.Status,
		&key.CreatedAt, &key.ExpiresAt,
	)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.LogDB(logger.DBLogParams{
				Ctx:          ctx,
				Operation:    logger.DBSelect,
				Table:        "idempotency_keys",
				Duration:     duration,
				RowsAffected: rowsAffected,
				Err:          nil,
				Custom:       map[string]any{"method": "find_by_key", "key": keyStr},
			})
			return nil, nil
		}
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "idempotency_keys",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
			Custom:       map[string]any{"method": "find_by_key", "key": keyStr},
		})
		return nil, err
	}

	rowsAffected = 1
	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "idempotency_keys",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
		Custom:       map[string]any{"method": "find_by_key", "key": keyStr},
	})
	key.ResponseBody = responseBody
	return &key, nil
}

func (r *PostgresIdempotencyKeyRepository) Delete(ctx context.Context, keyStr string) error {
	start := time.Now()
	tag, err := r.pool.Exec(ctx, "DELETE FROM idempotency_keys WHERE key = $1", keyStr)
	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = tag.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBDelete,
		Table:        "idempotency_keys",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
		Custom:       map[string]any{"method": "delete", "key": keyStr},
	})
	return err
}
