-- [WARNING] VARCHAR 길이 확장 (길이 증가 시 Rewrite 없음)
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(150);
