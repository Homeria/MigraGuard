-- [DANGER] 기존 데이터 전수 검사 (Access Exclusive Lock)
ALTER TABLE products ALTER COLUMN category SET NOT NULL;
