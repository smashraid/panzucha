-- 000001_initial_schema.down.sql

-- Drop triggers first
DROP TRIGGER IF EXISTS set_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS set_products_updated_at ON products;
DROP TRIGGER IF EXISTS set_users_updated_at ON users;

-- Drop functions
DROP FUNCTION IF EXISTS trigger_set_updated_at();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS inbox CASCADE;
DROP TABLE IF EXISTS outbox CASCADE;
DROP TABLE IF EXISTS idempotency_keys CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS users CASCADE;
