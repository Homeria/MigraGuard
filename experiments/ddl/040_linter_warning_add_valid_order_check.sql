-- Deliberately validates immediately so the linter flags the blocking scan.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE orders
    ADD CONSTRAINT check_orders_amount_immediate
    CHECK (total_amount >= 0);
COMMIT;
