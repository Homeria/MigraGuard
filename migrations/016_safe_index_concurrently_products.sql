-- [SAFE] 카테고리 기반 비차단 인덱스
CREATE INDEX CONCURRENTLY idx_products_category_search ON products(category);
