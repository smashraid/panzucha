package inbox

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Inbox struct {
	EventID     string
	PublishedAt time.Time
}

type InboxRepository interface {
	Create(ctx context.Context, tx pgx.Tx, eventID string) error
}
