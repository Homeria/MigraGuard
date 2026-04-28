-- [SAFE] Add a simple nullable column for marketing tracking.
-- Metadata-only change, does not block reads or writes.
ALTER TABLE orders ADD COLUMN promo_id VARCHAR(20);
