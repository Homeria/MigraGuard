-- [DANGER] Add a foreign key without 'NOT VALID'.
-- Forces a full table scan of the massive orders table to validate all existing IDs.
ALTER TABLE order_event_logs ADD CONSTRAINT fk_order_logs FOREIGN KEY (order_id) REFERENCES orders(id);
