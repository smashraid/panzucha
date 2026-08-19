package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"

	"panzucha/internal/shared/db"
	shareddomain "panzucha/internal/shared/domain"
	"panzucha/internal/shared/inbox"
	"panzucha/internal/shared/messaging"
)

// HandlerFunc processes a domain event payload within an active DB transaction.
// The transaction is committed only if the handler returns nil — any error
// rolls it back and routes the message to the DLQ.
type HandlerFunc func(ctx context.Context, tx pgx.Tx, payload []byte) error

// Consumer drains a queue using the transactional inbox pattern:
// begin tx → inbox dedup insert → handler → commit → ack.
//
// Transient DB errors (begin/inbox/commit) nack with requeue=true so the
// message is retried; only handler errors nack with requeue=false, routing
// the message to the dead-letter queue configured on the queue.
type Consumer struct {
	sub        messaging.Subscriber
	transactor db.Transactor
	inboxRepo  inbox.InboxRepository
	queue      messaging.QueueSpec
	handler    HandlerFunc
}

func New(
	sub messaging.Subscriber,
	transactor db.Transactor,
	inboxRepo inbox.InboxRepository,
	queue messaging.QueueSpec,
	handler HandlerFunc,
) *Consumer {
	return &Consumer{
		sub:        sub,
		transactor: transactor,
		inboxRepo:  inboxRepo,
		queue:      queue,
		handler:    handler,
	}
}

// Start subscribes to the queue and processes deliveries until ctx is
// cancelled or the deliveries channel closes.
func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.sub.Subscribe(ctx, c.queue, 10)
	if err != nil {
		return fmt.Errorf("subscribe %q: %w", c.queue.Name, err)
	}

	slog.Info("consumer: started", "queue", c.queue.Name)

	for {
		select {
		case <-ctx.Done():
			slog.Info("consumer: stopping", "queue", c.queue.Name)
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("deliveries channel closed for queue %q", c.queue.Name)
			}
			c.handleDelivery(ctx, d)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery) {
	eventID := d.MessageId
	if eventID == "" {
		eventID = d.CorrelationId
	}

	tx, err := c.transactor.BeginTx(ctx)
	if err != nil {
		slog.Error("consumer: begin tx failed", "err", err, "queue", c.queue.Name, "event_id", eventID)
		_ = d.Nack(false, true) // transient DB error — retry
		return
	}
	defer tx.Rollback(ctx) // no-op after successful Commit

	// 1. Inbox deduplication check
	if eventID != "" {
		err = c.inboxRepo.Create(ctx, tx, eventID)
		if errors.Is(err, shareddomain.ErrConflict) {
			slog.Info("consumer: duplicate event skipped", "queue", c.queue.Name, "event_id", eventID)
			_ = tx.Commit(ctx)
			_ = d.Ack(false)
			return
		}
		if err != nil {
			slog.Error("consumer: inbox insert failed", "err", err, "queue", c.queue.Name, "event_id", eventID)
			_ = d.Nack(false, true) // transient DB error — retry
			return
		}
	}

	// 2. Invoke business logic handler
	if err := c.handler(ctx, tx, d.Body); err != nil {
		slog.Error("consumer: handler error, routing to DLQ", "err", err, "queue", c.queue.Name, "event_id", eventID)
		_ = d.Nack(false, false) // requeue=false → dead-letter queue
		return
	}

	// 3. Commit transaction, then acknowledge
	if err := tx.Commit(ctx); err != nil {
		slog.Error("consumer: commit failed", "err", err, "queue", c.queue.Name, "event_id", eventID)
		_ = d.Nack(false, true) // transient DB error — retry
		return
	}

	_ = d.Ack(false)
}
