-- [WARNING] Rename a column in products.
-- Requires Access Exclusive Lock. Short duration but can block high-frequency reads.
ALTER TABLE products RENAME COLUMN category TO prod_category;
