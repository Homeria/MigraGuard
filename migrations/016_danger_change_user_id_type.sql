-- [DANGER] Change the type of a foreign key column.
-- Extremely heavy as it impacts both orders and users tables and requires rewrite.
ALTER TABLE orders ALTER COLUMN user_id TYPE BIGINT;
