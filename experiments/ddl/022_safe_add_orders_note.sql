-- Lint-clean nullable column addition for the orders table.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS deployment_note TEXT;
