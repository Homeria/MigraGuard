-- Deliberately omits CONCURRENTLY so the linter flags blocking index creation.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE INDEX IF NOT EXISTS idx_orders_total_amount_standard
    ON orders(total_amount);
