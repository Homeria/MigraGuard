-- [SAFE] Add a biography column to users.
-- Safe operation on a non-congested table.
ALTER TABLE users ADD COLUMN bio TEXT;
