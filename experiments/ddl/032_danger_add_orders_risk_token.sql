-- Lint-clean metadata change that becomes dangerous on a hot orders table.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS risk_review_token TEXT;
