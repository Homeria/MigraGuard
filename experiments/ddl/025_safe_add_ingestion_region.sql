-- Lint-clean nullable column addition for log ingestion metadata.
SET lock_timeout = '1s';
SET statement_timeout = '5s';
ALTER TABLE order_event_logs
    ADD COLUMN IF NOT EXISTS ingestion_region TEXT;
