package repositories

import (
	"context"
	"errors"
	"panzucha/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ domain.UserRepository = (*PostgresUserRepository)(nil)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
		SELECT id, email, name, password, role, created_at, updated_at
		FROM   users
		WHERE  id = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role,
		&u.Audit.CreatedAt, &u.Audit.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, name, password, role, created_at, updated_at
		FROM   users
		WHERE  email = $1`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role,
		&u.Audit.CreatedAt, &u.Audit.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	const q = `
    INSERT INTO users 
		(id, email, name, password, role, created_at, updated_at)
    VALUES 
		($1, $2, $3, $4, $5, NOW(), NOW())
    RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, q,
		u.ID, u.Email, u.Name, u.PasswordHash, u.Role,
	).Scan(&u.Audit.CreatedAt, &u.Audit.UpdatedAt)
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	const q = `
		UPDATE users 
		SET email		= $1, 
			name		= $2, 
			role		= $3, 
			updated_at	= NOW()
		WHERE id = $4
		RETURNING updated_at`
	err := r.pool.QueryRow(ctx, q,
		u.Email, u.Name, u.Role, u.ID,
	).Scan(&u.Audit.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
