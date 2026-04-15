-- [SAFE] 상수 기본값과 함께 컬럼 추가 (Postgres 11+ 최적화)
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT true;
