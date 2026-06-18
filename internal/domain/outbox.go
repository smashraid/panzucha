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
	Create(ctx context.Context, tx pgx.Tx, outbox Outbox) error
	List(ctx context.Context, limit int) ([]Outbox, error)
	MarkPublished(ctx context.Context, id string) error
}
