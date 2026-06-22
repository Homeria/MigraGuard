\set order_id random(1, 100000)
\set user_id random(1, 1000)

BEGIN;
SELECT id, status, total_amount
  FROM orders
 WHERE id = :order_id;
UPDATE orders
   SET user_id = :user_id
 WHERE id = :order_id;
COMMIT;
