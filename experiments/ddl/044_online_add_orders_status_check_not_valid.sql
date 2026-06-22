-- Staged constraint addition without immediate validation.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE orders
    ADD CONSTRAINT check_orders_status_valid
    CHECK (status IN ('PENDING', 'PAID', 'SHIPPING', 'CANCELLED')) NOT VALID;
COMMIT;
