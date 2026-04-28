-- [DANGER] Change column type of order_no in orders table.
-- Requires a full table rewrite, locking the massive orders table for a long duration.
ALTER TABLE orders ALTER COLUMN order_no TYPE TEXT;
