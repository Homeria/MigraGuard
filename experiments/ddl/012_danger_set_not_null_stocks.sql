-- [DANGER] Set a NOT NULL constraint on a column in inventory_stocks.
-- Requires a full table scan to validate existing rows, locking the table.
ALTER TABLE inventory_stocks ALTER COLUMN warehouse_id SET NOT NULL;
