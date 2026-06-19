-- Comparison fixture: statically safe migration that can still be risky under load.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_logs_created_at_async
    ON order_event_logs(created_at);
