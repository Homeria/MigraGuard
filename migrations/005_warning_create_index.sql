-- [WARNING] 일반 인덱스 생성 (쓰기 잠금 발생)
-- 리스크 점수: 중간 (Warning) - 트래픽이 있을 경우 점수 상승
CREATE INDEX idx_products_price_search ON products(price);
