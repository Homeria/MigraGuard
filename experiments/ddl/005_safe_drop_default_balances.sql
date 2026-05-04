-- [SAFE] Drop a default constraint from account_balances.
-- Metadata-only change, safe to execute.
ALTER TABLE account_balances ALTER COLUMN currency DROP DEFAULT;
