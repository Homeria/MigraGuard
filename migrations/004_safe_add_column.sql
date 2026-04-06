-- [SAFE] 단순 컬럼 추가 (Postgres 11+ 메타데이터만 변경)
-- 리스크 점수: 매우 낮음 (Safe)
ALTER TABLE users ADD COLUMN phone_number VARCHAR(20);
