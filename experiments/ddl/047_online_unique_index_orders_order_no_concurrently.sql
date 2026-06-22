-- Online preparation for a uniqueness guarantee using a concurrent unique index.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_order_no_unique_online
    ON orders(order_no);
