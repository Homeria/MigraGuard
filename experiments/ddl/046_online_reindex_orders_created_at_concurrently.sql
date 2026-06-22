-- DBMS-level online index maintenance used for forecast recommendation validation.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
REINDEX INDEX CONCURRENTLY idx_orders_created_at;
