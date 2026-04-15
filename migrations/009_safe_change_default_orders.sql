-- [SAFE] 기본값 설정 변경 (메타데이터만 변경)
ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'pending_review';
