\set product_id random(1, 500)

BEGIN;
UPDATE inventory_stocks
   SET updated_at = clock_timestamp()
 WHERE product_id = :product_id;
SELECT stock_quantity, reserved_quantity
  FROM inventory_stocks
 WHERE product_id = :product_id;
COMMIT;
