-- Deliberately sets NOT NULL directly so the linter flags the table scan.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE order_event_logs
    ALTER COLUMN event_type SET NOT NULL;
COMMIT;
