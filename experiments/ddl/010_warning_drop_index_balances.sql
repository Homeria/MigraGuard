-- [WARNING] Drop an index that might be in use.
-- Can cause performance degradation if queries rely on this index.
DROP INDEX IF EXISTS idx_order_logs_order_id;
