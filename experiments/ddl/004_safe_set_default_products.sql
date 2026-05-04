-- [SAFE] Set a default value for an existing column.
-- Does not require rewriting existing rows in modern PostgreSQL.
ALTER TABLE products ALTER COLUMN description SET DEFAULT 'No description available';
