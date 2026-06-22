\set user_id random(1, 1000)

BEGIN;
UPDATE account_balances
   SET last_updated_at = clock_timestamp(),
       version = version + 1
 WHERE user_id = :user_id;
SELECT balance, version
  FROM account_balances
 WHERE user_id = :user_id;
COMMIT;
