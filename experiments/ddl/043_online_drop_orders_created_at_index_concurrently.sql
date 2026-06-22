-- DBMS-level online index removal used for forecast recommendation validation.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
DROP INDEX CONCURRENTLY IF EXISTS idx_orders_created_at;
