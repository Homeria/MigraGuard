-- [DANGER] Complex multi-operation on account_balances.
-- Combining multiple alters on a hot-spot table creates long blocking chains.
ALTER TABLE account_balances ADD COLUMN frozen_at TIMESTAMP, ALTER COLUMN balance SET NOT NULL;
