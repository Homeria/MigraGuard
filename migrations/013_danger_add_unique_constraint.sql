-- [DANGER] Add a UNIQUE constraint.
-- Implicitly creates an index and locks the table to validate uniqueness across all rows.
ALTER TABLE users ADD CONSTRAINT unique_email_alt UNIQUE (email);
