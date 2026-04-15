-- [SAFE] 비차단 인덱스 생성 (Lock Level 4)
CREATE INDEX CONCURRENTLY idx_orders_created_at_ts ON orders(created_at);
