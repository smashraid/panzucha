-- outbox: written inside the same transaction as the business operation.
-- The relay worker polls WHERE published_at IS NULL and marks rows after
-- confirming broker delivery.
CREATE TABLE IF NOT EXISTS outbox (
    id           UUID        PRIMARY KEY,
    event_id     UUID        NOT NULL UNIQUE, -- consumer-side dedup key
    event_type   TEXT        NOT NULL,         -- routing key: "order.created" etc.
    payload      JSONB       NOT NULL,
    published_at TIMESTAMPTZ,                  -- NULL = pending, timestamp = delivered
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for the relay worker's polling query.
-- Partial index on unpublished rows only — keeps the index small and fast
-- as the table grows (published rows are the majority over time).
CREATE INDEX idx_outbox_unpublished
    ON outbox (created_at)
    WHERE published_at IS NULL;

-- inbox: consumer-side deduplication table.
-- The event_id PRIMARY KEY is the dedup guard — no extra logic needed.
-- INSERT ON CONFLICT DO NOTHING + RowsAffected == 0 = duplicate detected.
CREATE TABLE IF NOT EXISTS inbox (
    event_id     UUID        PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
