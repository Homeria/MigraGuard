-- [Warning] Lock competition (AccessExclusiveLock without rewrite)
CREATE INDEX idx_users_name ON users(name);
