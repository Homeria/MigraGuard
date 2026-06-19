-- Lint-clean concurrent index that is still dangerous on a saturated log table.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_logs_event_created_lookup
    ON order_event_logs(event_type, created_at DESC);
