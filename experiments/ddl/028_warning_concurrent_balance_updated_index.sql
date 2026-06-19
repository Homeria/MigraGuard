-- Lint-clean concurrent descending index for reconciliation scans.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_balances_last_updated_desc
    ON account_balances(last_updated_at DESC);
