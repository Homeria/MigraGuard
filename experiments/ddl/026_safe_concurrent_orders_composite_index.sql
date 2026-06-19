-- Lint-clean concurrent composite index for order lookups.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_status_created_at
    ON orders(status, created_at);
