-- [WARNING] 제약 조건 추가 (검증 제외 옵션, Lock Level 5)
ALTER TABLE products ADD CONSTRAINT price_check CHECK (price > 0) NOT VALID;
