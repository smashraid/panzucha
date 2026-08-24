package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"panzucha/internal/order/domain"
)

// HandleOrderCreated is the Phase A no-op consumer for the order.created
// queue. It logs the event and returns nil — the transactional inbox in the
// shared consumer loop handles dedup, commit, and ack. Malformed payloads
// return an error, which the consumer loop converts to a DLQ nack.
func HandleOrderCreated(ctx context.Context, tx pgx.Tx, payload []byte) error {
	var event domain.OrderCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal order.created event: %w", err)
	}

	slog.InfoContext(ctx, "consumer received order.created",
		"order_id", event.OrderID,
		"event_id", event.EventID,
		"user_id", event.UserID,
		"items_count", len(event.Items),
		"total_amount", event.TotalAmount,
	)
	return nil
}
