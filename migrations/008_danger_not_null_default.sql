-- [DANGER] 기본값이 포함된 NOT NULL 컬럼 추가
-- 리스크 점수: 높음 (Danger) - 모든 레코드에 기본값을 채워넣어야 함
ALTER TABLE products ADD COLUMN discount_active BOOLEAN NOT NULL DEFAULT false;
