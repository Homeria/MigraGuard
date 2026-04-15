-- [DANGER] 대규모 테이블 타입 변경 (Full Rewrite, Lock Level 8)
ALTER TABLE orders ALTER COLUMN total_amount TYPE NUMERIC(25,5);
