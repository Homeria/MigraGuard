-- [SAFE] 기본값 제거 (메타데이터)
ALTER TABLE users ALTER COLUMN created_at DROP DEFAULT;
