package repositories

import (
	"context"
	"database/sql"
	"errors"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ domain.UserRepository = (*PostgresUserRepository)(nil)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	if u.ID == "" {
		u.ID = domain.NewUserID()
	}
	query := `
        INSERT INTO users (id, email, name, password, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `
	_, err := r.pool.Exec(ctx, query, u.ID, u.Email, u.Name, u.Password)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, name, password, created_at, updated_at FROM users WHERE id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.Password,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, name, password, created_at, updated_at FROM users WHERE email = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.Password,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
        UPDATE users
        SET email = $1, name = $2, updated_at = NOW()
        WHERE id = $3
    `
	_, err := r.pool.Exec(ctx, query, u.Email, u.Name, u.ID)
	return err
}
