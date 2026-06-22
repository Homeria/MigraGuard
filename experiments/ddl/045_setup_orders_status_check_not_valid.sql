-- Precondition for VALIDATE CONSTRAINT benchmark cases.
ALTER TABLE orders
    ADD CONSTRAINT check_orders_status_valid
    CHECK (status IN ('PENDING', 'PAID', 'SHIPPING', 'CANCELLED')) NOT VALID;
