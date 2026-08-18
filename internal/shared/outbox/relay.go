package outbox

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"panzucha/internal/shared/messaging"
)

type Config struct {
	Interval    time.Duration
	BatchSize   int
	Concurrency int
	MaxRetries  int
}

func (c *Config) withDefaults() Config {
	if c.Interval <= 0 {
		c.Interval = 5 * time.Second
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 50
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 10
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = 10
	}
	return *c
}

// Relay design — single locked batch fetch, concurrent in-memory publish,
// two bulk writes:
//
//  1. ONE transaction locks up to BatchSize rows with FOR UPDATE SKIP LOCKED
//     in a single round-trip (ListAndLock).
//  2. Publishing happens CONCURRENTLY across goroutines (bounded by
//     Concurrency) but entirely in memory against the broker — no DB calls
//     during this phase, so the lock duration is just network time to
//     RabbitMQ, not N sequential round-trips.
//  3. Results are collected into two slices: succeeded IDs and failed IDs.
//  4. Two bulk UPDATEs (MarkPublishedBatch, MarkFailedBatch) close out the
//     SAME transaction — one round-trip each regardless of batch size.
//  5. Commit releases all locks at once.
//
// This means failure isolation no longer depends on the DB transaction at
// all — a failed publish for row 30 simply lands in the "failed" bucket and
// gets retried next cycle via retry_count; it never affects whether row 1-29
// get marked published, because that decision is made in Go, in memory,
// before either bulk UPDATE runs.
type Relay struct {
	pool       *pgxpool.Pool
	outboxRepo OutboxRepository
	broker     messaging.Broker
	cfg        Config
}

func NewRelay(pool *pgxpool.Pool, repo OutboxRepository, broker messaging.Broker, cfg Config) *Relay {
	return &Relay{pool: pool, outboxRepo: repo, broker: broker, cfg: cfg.withDefaults()}
}

func (r *Relay) Start(ctx context.Context) {
	slog.Info("outbox relay: started",
		"interval", r.cfg.Interval,
		"batch_size", r.cfg.BatchSize,
		"concurrency", r.cfg.Concurrency,
	)

	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()

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

func (r *Relay) runCycle(ctx context.Context) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "outbox relay: begin tx failed", "err", err)
		return
	}
	defer tx.Rollback(ctx) // no-op once Commit succeeds

	rows, err := r.outboxRepo.ListAndLock(ctx, tx, r.cfg.BatchSize, r.cfg.MaxRetries)
	if err != nil {
		slog.ErrorContext(ctx, "outbox relay: list failed", "err", err)
		return
	}
	if len(rows) == 0 {
		return // commit not needed — defer Rollback closes the empty tx
	}

	succeeded, failed := r.publishAll(ctx, rows)

	if err := r.outboxRepo.MarkPublishedBatch(ctx, tx, succeeded); err != nil {
		slog.ErrorContext(ctx, "outbox relay: mark published batch failed", "err", err)
		return
	}
	if err := r.outboxRepo.MarkFailedBatch(ctx, tx, failed, "publish failed — see logs for detail"); err != nil {
		slog.ErrorContext(ctx, "outbox relay: mark failed batch failed", "err", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "outbox relay: commit failed", "err", err)
		return
	}

	slog.InfoContext(ctx, "outbox relay: cycle complete",
		"fetched", len(rows),
		"published", len(succeeded),
		"failed", len(failed),
	)
}

// publishAll publishes every row concurrently (bounded by Concurrency) and
// returns two slices: IDs that published successfully, and IDs that failed.
// No DB calls happen here — purely broker I/O — which is what keeps the
// transaction's lock duration short despite a large batch.
func (r *Relay) publishAll(ctx context.Context, rows []Outbox) (succeeded, failed []string) {
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.cfg.Concurrency)

	for _, row := range rows {
		row := row // capture loop variable
		g.Go(func() error {
			err := r.broker.Publish(gctx, row.EventType, row.Payload)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				slog.WarnContext(ctx, "outbox relay: publish failed, will retry",
					"outbox_id", row.ID,
					"event_type", row.EventType,
					"retry_count", row.RetryCount+1,
					"err", err,
				)
				failed = append(failed, row.ID)
			} else {
				succeeded = append(succeeded, row.ID)
			}
			return nil // never return the error — one bad publish must not
			// cancel gctx and abort publishing for the rest of the batch
		})
	}

	_ = g.Wait()
	return succeeded, failed
}
