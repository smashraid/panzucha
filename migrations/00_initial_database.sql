-- 00_initial_database.sql
-- Production-grade schema aligned with Go domain entities.
-- Run once. Idempotent by design (no IF NOT EXISTS for core tables to fail fast on drift).

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─────────────────────────────────────────────────────────────────────────────
-- USERS
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    TEXT,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by    TEXT
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- ─────────────────────────────────────────────────────────────────────────────
-- PRODUCTS
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE products (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    price       DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock       INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    version     INT NOT NULL DEFAULT 1, -- For optimistic concurrency control
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  TEXT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by  TEXT
);

CREATE INDEX idx_products_name ON products(name);

-- ─────────────────────────────────────────────────────────────────────────────
-- ORDERS (Header)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE orders (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by   TEXT,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by   TEXT
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- ORDER ITEMS (1:N Relationship)
-- Normalized from []OrderItem slice. Relational DBs require a child table.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE order_items (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id   UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- IDEMPOTENCY KEYS
-- ResponseBody mapped to JSONB for structured API responses. 
-- Use BYTEA if you strictly need raw bytes.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE idempotency_keys (
    id              BIGSERIAL PRIMARY KEY,
    key             TEXT UNIQUE NOT NULL,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     UUID,
    response_status INT,
    response_body   JSONB,
    status          VARCHAR(20) NOT NULL DEFAULT 'processing',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_idempotency_keys_key ON idempotency_keys(key);
CREATE INDEX idx_idempotency_keys_status_resource ON idempotency_keys(status, resource_type);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

-- ─────────────────────────────────────────────────────────────────────────────
-- AUTOMATED AUDIT TRIGGERS
-- Keeps Audit.UpdatedAt in sync without relying on app logic.
-- Optional but strongly recommended for production consistency.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER set_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();