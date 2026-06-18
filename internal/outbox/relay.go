package outbox

import (
	"context"
	"log/slog"
	"time"

	"panzucha/internal/domain"
	"panzucha/internal/messaging"

	"github.com/jackc/pgx/v5/pgxpool"
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
	pool       *pgxpool.Pool
	cfg        Config
}

// NewRelay creates a Relay. Start() must be called explicitly — typically
// as a goroutine in main.go after the broker connection is established.
func NewRelay(repo domain.OutboxRepository, broker messaging.Broker, pool *pgxpool.Pool, cfg Config) *Relay {
	return &Relay{
		outboxRepo: repo,
		broker:     broker,
		pool:       pool,
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
	// Fetch a batch to know how many rows are pending —
	// then process each in its own tx to minimise lock duration.
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "outbox relay: begin tx failed", "err", err)
		return
	}
	defer tx.Rollback(ctx)

	rows, err := r.outboxRepo.ListAndLock(ctx, tx, r.cfg.BatchSize)
	if err != nil {
		slog.ErrorContext(ctx, "outbox relay: list failed", "err", err)
		return
	}
	if len(rows) == 0 {
		tx.Rollback(ctx)
		return
	}

	for _, row := range rows {
		if err := r.broker.Publish(ctx, row.EventType, row.Payload); err != nil {
			slog.ErrorContext(ctx, "outbox relay: publish failed",
				"err", err, "outbox_id", row.ID, "event_type", row.EventType)
			// Do not mark published — leave for next cycle.
			// Roll back the whole batch so no row is partially marked.
			tx.Rollback(ctx)
			return
		}
		if err := r.outboxRepo.MarkPublished(ctx, tx, row.ID); err != nil {
			slog.ErrorContext(ctx, "outbox relay: mark published failed",
				"err", err, "outbox_id", row.ID)
			tx.Rollback(ctx)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "outbox relay: commit failed", "err", err)
	}
}
