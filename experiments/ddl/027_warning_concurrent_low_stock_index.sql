-- Lint-clean concurrent partial index for low-stock searches.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inventory_low_stock
    ON inventory_stocks(product_id)
    WHERE stock_quantity <= 10;
