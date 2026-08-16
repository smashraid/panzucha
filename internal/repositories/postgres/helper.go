package postgres

import (
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation returns true when the error is a Postgres unique
// constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// unmarshalJSON decodes JSON bytes into v.
// Centralised here so all repo JSON handling goes through one place.
func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
