# 🏦 Fintech Commerce Schema Design (v3.8)

This document defines the high-concurrency, high-volume database schema designed to stress-test MigraGuard's risk engine.

---

## 1. Schema Overview

The schema is divided into three functional zones to simulate different types of database stress.

### A. Hot Spot (High Contention)
- **Table:** `account_balances`
- **Purpose:** Simulate extreme row-level lock contention.
- **Scenario:** Thousands of concurrent `UPDATE` transactions during a DDL execution.

### B. Heavy Body (Massive Volume)
- **Table:** `orders`
- **Purpose:** Simulate long-running DDLs (Table Rewrite).
- **Scenario:** Changing column types or adding non-null constraints on 100GB+ datasets.

### C. Modern Payload (Complex Types)
- **Table:** `order_event_logs`
- **Purpose:** Simulate I/O and CPU-bound index creation.
- **Scenario:** Indexing large `JSONB` fields or adding complex `CHECK` constraints.

---

## 2. Table Definitions (SQL)

### 2.1. Account Balances (Hot Spot)
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

### 2.2. Orders (Heavy Body)
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

### 2.3. Inventory Stocks (Contention Point)
```sql
CREATE TABLE inventory_stocks (
    product_id BIGINT PRIMARY KEY,
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    reserved_quantity INT DEFAULT 0,
    warehouse_id INT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2.4. Order Event Logs (Massive Insert)
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

## 3. Research Scenarios

| Scenario | Target Table | DDL Operation | Expected Risk |
| :--- | :--- | :--- | :--- |
| **Flash Sale** | `inventory_stocks` | `ADD COLUMN warehouse_id` | High `C_peak` due to row-level blocking. |
| **Data Migration** | `orders` | `ALTER COLUMN order_no TYPE TEXT` | High `T_ddl` due to table rewrite. |
| **Audit Compliance** | `order_event_logs` | `CREATE INDEX idx_payload` | High I/O impact on concurrent `INSERT`s. |
