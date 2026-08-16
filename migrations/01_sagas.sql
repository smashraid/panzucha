CREATE TABLE outbox (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id     UUID        NOT NULL UNIQUE,
    event_type   TEXT        NOT NULL,
    payload      JSONB       NOT NULL,
    published_at TIMESTAMPTZ,            -- NULL = pending
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inbox (
    event_id     UUID        PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);