-- Lint-clean concurrent partial index for payment event searches.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_logs_payment_created_at
    ON order_event_logs(event_type, created_at)
    WHERE event_type = 'PAYMENT_SUCCESS';
