-- Lint-clean nullable column addition for balance reconciliation.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
ALTER TABLE account_balances
    ADD COLUMN IF NOT EXISTS reconciliation_reference TEXT;
