-- [DANGER] Truncate a large log table.
-- Irreversible and takes an Access Exclusive Lock, stopping all logging activity.
TRUNCATE TABLE order_event_logs;
