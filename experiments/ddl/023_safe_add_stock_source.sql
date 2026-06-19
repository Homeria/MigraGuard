-- Lint-clean nullable column addition for inventory metadata.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
ALTER TABLE inventory_stocks
    ADD COLUMN IF NOT EXISTS stock_source TEXT;
