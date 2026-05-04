-- [DANGER] Add a complex check constraint on a heavy table.
-- Triggers a full scan of the logs table, blocking inserts.
ALTER TABLE order_event_logs ADD CONSTRAINT check_event_type CHECK (event_type IS NOT NULL);
