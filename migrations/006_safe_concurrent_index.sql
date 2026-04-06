-- [SAFE/WARNING] 컨커런트 인덱스 생성 (락 최소화)
-- 리스크 점수: 낮음 (Safe) - MigraGuard가 CONCURRENTLY를 인식하는지 테스트
CREATE INDEX CONCURRENTLY idx_orders_created_at_search ON orders(created_at);
