-- [WARNING] Add a check constraint but skip validation.
-- Safe for now, but requires a separate 'VALIDATE CONSTRAINT' step later.
ALTER TABLE orders ADD CONSTRAINT check_total_positive CHECK (total_amount > 0) NOT VALID;
