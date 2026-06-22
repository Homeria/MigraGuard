-- Deferred validation step after adding a NOT VALID constraint.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE orders VALIDATE CONSTRAINT check_orders_status_valid;
COMMIT;
