-- [SAFE] Create an index concurrently on logs to avoid blocking inserts.
-- Background process, safe for high-traffic log tables.
CREATE INDEX CONCURRENTLY idx_logs_created_at_async ON order_event_logs(created_at);
