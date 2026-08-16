-- orders stores the header — no items column, no JSONB.
CREATE TABLE IF NOT EXISTS orders (
    id           UUID          PRIMARY KEY,
    user_id      UUID          NOT NULL REFERENCES users(id),
    status       TEXT          NOT NULL DEFAULT 'pending'
                               CHECK (status IN ('pending','confirmed','shipped','cancelled')),
    total_amount NUMERIC(12,2) NOT NULL CHECK (total_amount >= 0),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_by   UUID          NOT NULL,
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_by   UUID          NOT NULL
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status  ON orders(status);

-- order_items stores line items — FK to orders with CASCADE DELETE.
-- If an order is deleted, its items are deleted automatically.
CREATE TABLE IF NOT EXISTS order_items (
    id         UUID          PRIMARY KEY,
    order_id   UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID          NOT NULL REFERENCES products(id),
    quantity   INT           NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- idempotency_keys expire after 24h.
-- expires_at in queries makes them logically invisible before cleanup runs.
CREATE TABLE IF NOT EXISTS idempotency_keys (
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

CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);
