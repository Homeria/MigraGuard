-- Lint-clean metadata change that becomes dangerous under inventory load.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
ALTER TABLE inventory_stocks
    ADD COLUMN IF NOT EXISTS restock_batch_reference TEXT;
