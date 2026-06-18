package outbox

import (
	"context"
	"log/slog"
	"time"

	"panzucha/internal/domain"
	"panzucha/internal/messaging"
)

// Config controls relay behaviour. Sensible defaults are applied by NewRelay
// so callers only need to override what they care about.
type Config struct {
	// Interval between polling cycles. Default: 5s.
	// Shorter = lower latency, higher DB load.
	// Longer = higher latency, lower DB load.
	Interval time.Duration

	// BatchSize is the maximum number of outbox rows fetched per cycle.
	// FOR UPDATE SKIP LOCKED means concurrent relays share the work —
	// each instance locks its own batch and skips rows locked by peers.
	// Default: 50.
	BatchSize int
}

func (c *Config) withDefaults() Config {
	if c.Interval <= 0 {
		c.Interval = 5 * time.Second
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 50
	}
	return *c
}

// Relay reads unpublished outbox rows, publishes them to the broker, and
// marks them as published. It runs as a goroutine started in main.go.
//
// At-least-once delivery guarantee:
//   - If the process crashes after Publish but before markPublished, the row
//     stays unpublished and the next relay cycle retries it.
//   - Consumers must be idempotent (inbox dedup) to handle the rare duplicate.
//
// Horizontal scaling:
//   - Multiple relay instances are safe because FOR UPDATE SKIP LOCKED ensures
//     each instance locks a disjoint set of rows. No double-publish under load.
type Relay struct {
	outboxRepo domain.OutboxRepository
	broker     messaging.Broker
	cfg        Config
}

// NewRelay creates a Relay. Start() must be called explicitly — typically
// as a goroutine in main.go after the broker connection is established.
func NewRelay(repo domain.OutboxRepository, broker messaging.Broker, cfg Config) *Relay {
	return &Relay{
		outboxRepo: repo,
		broker:     broker,
		cfg:        cfg.withDefaults(),
	}
}

// Start runs the relay loop until ctx is cancelled.
// Designed to run as: go relay.Start(ctx)
//
// On each tick:
//  1. Fetch a batch of unpublished rows (FOR UPDATE SKIP LOCKED).
//  2. Publish each row to the broker.
//  3. Mark each successfully published row with published_at = NOW().
//
// A failed publish is logged and skipped — the row stays unpublished
// and will be retried on the next cycle. This means a single bad row
// does not block the rest of the batch.
func (r *Relay) Start(ctx context.Context) {
	slog.Info("outbox relay: started",
		"interval", r.cfg.Interval,
		"batch_size", r.cfg.BatchSize,
	)

	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()

	// Run one cycle immediately on startup so there's no wait for the
	// first tick — useful after a restart where rows may already be pending.
	r.runCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox relay: stopped")
			return
		case <-ticker.C:
			r.runCycle(ctx)
		}
	}
}

// runCycle executes one fetch-publish-mark cycle.
// Errors inside the cycle are logged but never returned — the relay
// must keep running even if a single cycle fails.
func (r *Relay) runCycle(ctx context.Context) {
	rows, err := r.outboxRepo.List(ctx, r.cfg.BatchSize)
	if err != nil {
		slog.ErrorContext(ctx, "outbox relay: failed to fetch rows", "err", err)
		return
	}
	if len(rows) == 0 {
		return // nothing pending — common case, no log noise
	}

	slog.InfoContext(ctx, "outbox relay: processing batch", "count", len(rows))

	for _, row := range rows {
		r.publishRow(ctx, row)
	}
}

// publishRow publishes one outbox row and marks it published on success.
// Failures are logged individually so one bad row doesn't abort the batch.
func (r *Relay) publishRow(ctx context.Context, row domain.Outbox) {
	if err := r.broker.Publish(ctx, row.EventType, row.Payload); err != nil {
		// Publish failed — leave published_at as NULL so the next cycle retries.
		// Common causes: broker temporarily unavailable, network blip.
		// The reconnect goroutine in RabbitMQBroker handles connection recovery.
		slog.ErrorContext(ctx, "outbox relay: publish failed",
			"err", err,
			"outbox_id", row.ID,
			"event_type", row.EventType,
			"event_id", row.EventID,
		)
		return
	}

	// Mark published only after the broker confirms delivery.
	// If this update fails the row will be republished on the next cycle —
	// consumers handle duplicates via the inbox dedup pattern.
	if err := r.outboxRepo.MarkPublished(ctx, row.ID); err != nil {
		slog.ErrorContext(ctx, "outbox relay: failed to mark published",
			"err", err,
			"outbox_id", row.ID,
		)
		return
	}

	slog.InfoContext(ctx, "outbox relay: published",
		"outbox_id", row.ID,
		"event_type", row.EventType,
		"event_id", row.EventID,
	)
}
