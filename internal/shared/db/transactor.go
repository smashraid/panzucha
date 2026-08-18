package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transactor abstracts transaction lifecycle management.
// Defined here so service and consumer layers can depend on this interface
// instead of pgxpool.Pool directly — keeping pgx out of service packages
// and making them testable without a real database.
type Transactor interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

// PgxTransactor implements Transactor using pgxpool.
type PgxTransactor struct {
	pool *pgxpool.Pool
}

// NewPgxTransactor is the constructor. Accepts the shared pool that is
// created once in main.go and injected everywhere that needs it.
func NewPgxTransactor(pool *pgxpool.Pool) *PgxTransactor {
	return &PgxTransactor{pool: pool}
}

func (t *PgxTransactor) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return t.pool.Begin(ctx)
}
