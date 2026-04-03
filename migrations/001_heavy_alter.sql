-- migrations/001_heavy_alter.sql (테이블 전체를 다시 쓰는 무거운 작업)
ALTER TABLE pgbench_accounts ADD COLUMN bio text DEFAULT 'Hello, MigraGuard!';
