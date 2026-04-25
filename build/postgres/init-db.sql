-- 🛡️ MigraGuard Simulation Target Schema (E-commerce)

-- 1. 통계 수집 확장기 활성화
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 2. 사용자 테이블 (읽기 위주)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 3. 상품 테이블 (읽기 위주, 재고 업데이트 시 쓰기 발생)
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    stock_quantity INTEGER DEFAULT 0,
    category VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

-- 4. 주문 테이블 (쓰기 위주, 트래픽 피크 시 급증)
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total_amount DECIMAL(12,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, paid, shipped, cancelled
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

-- 5. 주문 상세 내역 (대량 쓰기 발생, JOIN 부하 유도)
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12,2) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- 6. 기초 데이터 삽입 (시뮬레이션 초기값 강화)
-- 사용자 1,000명
INSERT INTO users (username, email)
SELECT 
    'user_' || i, 
    'user_' || i || '@example.com'
FROM generate_series(1, 1000) s(i)
ON CONFLICT DO NOTHING;

-- 상품 100개
INSERT INTO products (name, price, stock_quantity, category)
SELECT 
    'Product ' || i, 
    (random() * 1000 + 10)::decimal(12,2), 
    1000, 
    (ARRAY['Electronics', 'Clothing', 'Books', 'Home'])[floor(random()*4)+1]
FROM generate_series(1, 100) s(i)
ON CONFLICT DO NOTHING;

-- 주문 100,000건 (테이블 크기를 키워 T_ddl을 늘림)
INSERT INTO orders (user_id, total_amount, status, created_at)
SELECT 
    floor(random() * 1000) + 1,
    (random() * 500 + 10)::decimal(12,2),
    'paid',
    now() - (random() * interval '7 days')
FROM generate_series(1, 100000) s(i);

-- 주문 상세 내역 200,000건
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
SELECT 
    floor(random() * 100000) + 1,
    floor(random() * 100) + 1,
    floor(random() * 3) + 1,
    100.0
FROM generate_series(1, 200000) s(i);
