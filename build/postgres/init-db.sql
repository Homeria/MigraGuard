-- 🛡️ MigraGuard Simulation Target Schema (Fintech Commerce)
-- Optimized for High-Concurrency Lock Contention & Heavy Data Analysis

-- 1. 확장기 활성화
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 2. 사용자 기본 테이블
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. [Hot Spot] 계좌 잔액 테이블 (고빈도 UPDATE 경합 유도)
CREATE TABLE IF NOT EXISTS account_balances (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    balance DECIMAL(19, 4) NOT NULL DEFAULT 0,
    point_balance INT NOT NULL DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'KRW',
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    version BIGINT NOT NULL DEFAULT 0
);

-- 4. 상품 정보 테이블
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    category VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. [Contention Point] 재고 관리 테이블 (Hot-Row 경합 유도)
CREATE TABLE IF NOT EXISTS inventory_stocks (
    product_id INTEGER PRIMARY KEY REFERENCES products(id),
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    reserved_quantity INT DEFAULT 0,
    warehouse_id INT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. [Heavy Body] 주문 메인 테이블 (대규모 데이터 누적)
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(50) UNIQUE NOT NULL,
    user_id INTEGER REFERENCES users(id),
    total_amount DECIMAL(19, 4) NOT NULL,
    status VARCHAR(20) NOT NULL, -- PENDING, PAID, SHIPPING, CANCELLED
    order_details JSONB, -- Modern Payload 부하 유도
    shipping_address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- 7. [Massive Log] 주문 이벤트 로그 (초거대 INSERT 전용 테이블)
CREATE TABLE IF NOT EXISTS order_event_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,
    event_type VARCHAR(50),
    raw_payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_order_logs_order_id ON order_event_logs(order_id);

-- ---------------------------------------------------------
-- 기초 데이터 시딩 (연구용 볼륨 확보)
-- ---------------------------------------------------------

-- 1,000 Users & Balances
INSERT INTO users (username, email)
SELECT 'user_' || i, 'user_' || i || '@example.com'
FROM generate_series(1, 1000) s(i);

INSERT INTO account_balances (user_id, balance, point_balance)
SELECT id, (random() * 1000000)::decimal(19,4), floor(random() * 1000)
FROM users;

-- 500 Products & Stocks
INSERT INTO products (name, price, category)
SELECT 'Product ' || i, (random() * 1000 + 10)::decimal(12,2), (ARRAY['Electronics', 'Clothing', 'Books', 'Home'])[floor(random()*4)+1]
FROM generate_series(1, 500) s(i);

INSERT INTO inventory_stocks (product_id, stock_quantity, warehouse_id)
SELECT id, 10000, floor(random() * 10) + 1
FROM products;

-- 100,000 Orders (Initial Volume)
INSERT INTO orders (order_no, user_id, total_amount, status, order_details)
SELECT 
    'ORD-' || i || '-' || floor(random()*100000),
    floor(random() * 1000) + 1,
    (random() * 500 + 10)::decimal(19,4),
    'PAID',
    '{"source": "mobile", "campaign": "spring_sale"}'::jsonb
FROM generate_series(1, 100000) s(i);

-- 200,000 Event Logs
INSERT INTO order_event_logs (order_id, event_type, raw_payload)
SELECT 
    floor(random() * 100000) + 1,
    (ARRAY['ORDER_CREATED', 'PAYMENT_COMPLETED', 'STOCK_RESERVED'])[floor(random()*3)+1],
    '{"ip": "127.0.0.1", "user_agent": "Mozilla/5.0"}'::jsonb
FROM generate_series(1, 200000) s(i);
