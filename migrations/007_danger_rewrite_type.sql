-- [DANGER] 컬럼 타입 변경 (Full Rewrite 발생)
-- 리스크 점수: 매우 높음 (Danger) - 테이블 전체를 다시 써야 함
ALTER TABLE orders ALTER COLUMN status TYPE TEXT;
