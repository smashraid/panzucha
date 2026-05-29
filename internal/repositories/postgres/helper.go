package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation returns true when the error is a Postgres unique
// constraint violation (SQLSTATE 23505). Centralised here so every repo
// can call it without duplicating the pgconn import and type assertion.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
