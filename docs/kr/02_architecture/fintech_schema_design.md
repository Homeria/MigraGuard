# 🏦 핀테크 커머스 스키마 설계 (v3.8)

이 문서는 MigraGuard의 리스크 엔진을 테스트하기 위해 설계된 고동시성, 대규모 데이터베이스 스키마를 정의합니다.

---

## 1. 스키마 개요

스키마는 서로 다른 유형의 데이터베이스 부하를 시뮬레이션하기 위해 세 가지 기능 영역으로 나뉩니다.

### A. 핫스팟 (고빈도 경합)
- **테이블:** `account_balances`
- **목적:** 극단적인 행 수준(Row-level) 락 경합 시뮬레이션.
- **시나리오:** DDL 실행 중 수천 건의 동시 `UPDATE` 트랜잭션 발생.

### B. 헤비 바디 (대규모 데이터)
- **테이블:** `orders`
- **목적:** 장시간 소요되는 DDL(Table Rewrite) 시뮬레이션.
- **시나리오:** 100GB 이상의 데이터셋에서 컬럼 타입 변경 또는 Non-null 제약 조건 추가.

### C. 모던 페이로드 (복합 타입)
- **테이블:** `order_event_logs`
- **목적:** I/O 및 CPU 바운드 인덱스 생성 시뮬레이션.
- **시나리오:** 거대한 `JSONB` 필드에 인덱스 생성 또는 복잡한 `CHECK` 제약 조건 추가.

---

## 2. 테이블 정의 (SQL)

### 2.1. 계좌 잔액 (Hot Spot)
```sql
CREATE TABLE account_balances (
    user_id BIGINT PRIMARY KEY,
    balance DECIMAL(19, 4) NOT NULL DEFAULT 0,
    point_balance INT NOT NULL DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'KRW',
    last_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version BIGINT NOT NULL DEFAULT 0
);
```

### 2.2. 주문 (Heavy Body)
```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(50) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(19, 4) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    order_details JSONB, 
    shipping_address TEXT
);
```

### 2.3. 재고 (Contention Point)
```sql
CREATE TABLE inventory_stocks (
    product_id BIGINT PRIMARY KEY,
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    reserved_quantity INT DEFAULT 0,
    warehouse_id INT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2.4. 주문 이벤트 로그 (Massive Insert)
```sql
CREATE TABLE order_event_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,
    event_type VARCHAR(50),
    raw_payload JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 3. 연구 시나리오

| 시나리오 | 대상 테이블 | DDL 작업 | 예상 리스크 |
| :--- | :--- | :--- | :--- |
| **플래시 세일** | `inventory_stocks` | `ADD COLUMN warehouse_id` | 행 수준 차단으로 인한 높은 `C_peak`. |
| **데이터 마이그레이션** | `orders` | `ALTER COLUMN order_no TYPE TEXT` | 테이블 재작성(Rewrite)으로 인한 높은 `T_ddl`. |
| **감사 준수** | `order_event_logs` | `CREATE INDEX idx_payload` | 동시 `INSERT` 작업에 대한 높은 I/O 영향. |
