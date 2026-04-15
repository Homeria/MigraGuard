-- [DANGER] 유니크 제약 조건 추가 (인덱스 생성 및 전체 검사)
ALTER TABLE order_items ADD CONSTRAINT unique_order_product UNIQUE (order_id, product_id);
