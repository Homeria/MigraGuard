-- Lint-clean default change that still needs an exclusive metadata lock.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE order_event_logs
    ALTER COLUMN event_type SET DEFAULT 'UNKNOWN';
COMMIT;
