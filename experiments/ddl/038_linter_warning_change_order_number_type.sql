-- Deliberately changes a column type so the linter flags a table rewrite.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE orders
    ALTER COLUMN order_no TYPE TEXT;
COMMIT;
