-- Lint-clean transactional default change.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE orders
    ALTER COLUMN status SET DEFAULT 'PENDING';
COMMIT;
