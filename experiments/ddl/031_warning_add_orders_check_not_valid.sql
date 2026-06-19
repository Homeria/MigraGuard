-- Lint-clean staged check constraint that defers the table scan.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE orders
    ADD CONSTRAINT check_orders_total_nonnegative
    CHECK (total_amount >= 0) NOT VALID;
COMMIT;
