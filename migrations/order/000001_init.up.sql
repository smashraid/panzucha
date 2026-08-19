CREATE TABLE users (
    id            UUID        PRIMARY KEY,
    name          TEXT        NOT NULL,
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    TEXT,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by    TEXT
);

CREATE TABLE products (
    id          UUID          PRIMARY KEY,
    name        TEXT          NOT NULL,
    description TEXT,
    price       NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    stock       INT           NOT NULL DEFAULT 0 CHECK (stock >= 0),
    version     INT           NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_by  TEXT,
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_by  TEXT
);

CREATE TABLE orders (
    id           UUID          PRIMARY KEY,
    user_id      UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status       TEXT          NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending','confirmed','shipped','cancelled')),
    total_amount NUMERIC(12,2) NOT NULL CHECK (total_amount >= 0),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_by   TEXT          NOT NULL,
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_by   TEXT          NOT NULL
);
CREATE INDEX idx_orders_user_created ON orders (user_id, created_at DESC);

CREATE TABLE order_items (
    id         UUID          PRIMARY KEY,
    order_id   UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID          NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INT           NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_items_order_id   ON order_items (order_id);
CREATE INDEX idx_order_items_product_id ON order_items (product_id);

CREATE TABLE idempotency_keys (
    key             TEXT        PRIMARY KEY,
    resource_type   TEXT        NOT NULL,
    resource_id     UUID,
    response_status INT,
    response_body   JSONB,
    status          TEXT        NOT NULL DEFAULT 'processing'
                                CHECK (status IN ('processing','completed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);

CREATE TABLE outbox (
    id           UUID        PRIMARY KEY,
    event_id     UUID        NOT NULL UNIQUE,
    event_type   TEXT        NOT NULL,
    payload      JSONB       NOT NULL,
    published_at TIMESTAMPTZ,
    retry_count  INT         NOT NULL DEFAULT 0,
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;

CREATE TABLE inbox (
    event_id     UUID        PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);