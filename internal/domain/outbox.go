package domain

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Outbox struct {
	ID          string
	EventID     string
	EventType   string
	Payload     []byte
	PublishedAt *time.Time
	CreatedAt   time.Time
}

type OutboxRepository interface {
	GetByID(ctx context.Context, id string) (*Outbox, error)
	Create(ctx context.Context, tx pgx.Tx, outbox Outbox) error
}
