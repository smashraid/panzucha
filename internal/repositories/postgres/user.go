package repositories

import (
	"context"
	"database/sql"
	"errors"
	"panzucha/internal/domain"
	"panzucha/internal/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ domain.UserRepository = (*PostgresUserRepository)(nil)

type PostgresUserRepository struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

func NewPostgresUserRepository(pool *pgxpool.Pool, log *logger.Logger) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool, logger: log}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	start := time.Now()
	if u.ID == "" {
		u.ID = domain.NewUserID()
	}
	query := `
        INSERT INTO users (id, email, name, password)
		VALUES ($1, $2, $3, $4)
    `
	result, err := r.pool.Exec(ctx, query, u.ID, u.Email, u.Name, u.Password)
	duration := time.Since(start)
	rowsAffected := int64(0)
	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBInsert,
		Table:        "users",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
	})
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	start := time.Now()
	query := `SELECT id, email, name, password, created_at, updated_at FROM users WHERE id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.Password,
		&u.CreatedAt, &u.UpdatedAt,
	)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if errors.Is(err, sql.ErrNoRows) {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "users",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          nil,
		})
		return nil, nil
	}

	if err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "users",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
		})
		return nil, err
	}

	rowsAffected = 1
	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "users",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
	})
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	start := time.Now()
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, password, created_at, updated_at FROM users WHERE email = $1`,
		email).Scan(
		&u.ID, &u.Email, &u.Name, &u.Password,
		&u.CreatedAt, &u.UpdatedAt,
	)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if errors.Is(err, sql.ErrNoRows) {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "users",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          nil,
		})
		return nil, nil
	}

	if err != nil {
		r.logger.LogDB(logger.DBLogParams{
			Ctx:          ctx,
			Operation:    logger.DBSelect,
			Table:        "users",
			Duration:     duration,
			RowsAffected: rowsAffected,
			Err:          err,
		})
		return nil, err
	}

	rowsAffected = 1
	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBSelect,
		Table:        "users",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          nil,
	})
	return &u, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	start := time.Now()
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET email=$1, name=$2, role=$3, updated_at=$4 WHERE id=$5`,
		u.Email, u.Name, u.Role, time.Now(), u.ID)
	duration := time.Since(start)
	rowsAffected := int64(0)

	if err == nil {
		rowsAffected = result.RowsAffected()
	}

	r.logger.LogDB(logger.DBLogParams{
		Ctx:          ctx,
		Operation:    logger.DBUpdate,
		Table:        "users",
		Duration:     duration,
		RowsAffected: rowsAffected,
		Err:          err,
	})
	return err
}
