-- Deliberately renames a column so the linter flags client compatibility risk.
BEGIN;
SET LOCAL lock_timeout = '1s';
SET LOCAL statement_timeout = '5s';
ALTER TABLE products
    RENAME COLUMN description TO product_description;
COMMIT;
