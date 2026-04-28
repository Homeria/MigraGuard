-- [WARNING] Increase the length of a VARCHAR column.
-- Usually metadata-only, but can trigger a rewrite in older PG versions or specific settings.
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(100);
