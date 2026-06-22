-- Deliberately sets NOT NULL directly on a nullable orders column.
-- This is used to test immediate validation risk on the benchmark workload table.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE orders
    ALTER COLUMN shipping_address SET NOT NULL;
COMMIT;
