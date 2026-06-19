-- Lint-clean staged constraint on a workload-sensitive inventory table.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE inventory_stocks
    ADD CONSTRAINT check_stock_quantity_ceiling
    CHECK (stock_quantity <= 1000000) NOT VALID;
COMMIT;
