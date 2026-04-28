# ⚙️ Fintech Workload Implementation Specifications

This document details the high-concurrency transaction model and table relationships implemented in v3.8-Phase 1 for research validation.

---

## 1. Transaction Flow (Complex Purchase)

The `load_generator` now simulates a real-world fintech transaction instead of simple single-table inserts. Each "Order Success" in the stats involves the following ACID transaction:

1.  **Inventory Check (`inventory_stocks`):**
    - `UPDATE inventory_stocks SET stock_quantity = stock_quantity - 1 WHERE product_id = ? AND stock_quantity >= 1`
    - *Purpose:* Create row-level lock contention on hot product items.
2.  **Balance Update (`account_balances`):**
    - `UPDATE account_balances SET balance = balance - price WHERE user_id = ? AND balance >= price`
    - *Purpose:* Simulate high-frequency updates on user account records.
3.  **Order Persistence (`orders`):**
    - `INSERT INTO orders (...) RETURNING id`
    - *Purpose:* Grow the main transaction table to increase the cost of table rewrites.
4.  **Audit Logging (`order_event_logs`):**
    - `INSERT INTO order_event_logs (...)`
    - *Purpose:* Simulate massive append-only logging traffic and I/O pressure.

---

## 2. Table Characteristics for Research

### A. account_balances (The Hot Spot)
- **Characteristic:** Small row size but extremely high update frequency.
- **Risk Scenario:** Adding a standard index or changing a column type on this table will block all financial transactions, leading to immediate system failure.

### B. orders (The Heavy Body)
- **Characteristic:** Massive data volume and complex indices.
- **Risk Scenario:** Operations that trigger a `Table Rewrite` (e.g., `VARCHAR` to `TEXT` or `SET NOT NULL`) will result in high `T_ddl` values, accurately predicted by MigraGuard.

### C. inventory_stocks (The Contention Point)
- **Characteristic:** Strict integrity constraints (`CHECK stock_quantity >= 0`).
- **Risk Scenario:** During a "Flash Sale" profile, this table becomes the bottleneck. Adding a column or modifying constraints here will cause a massive queue in the connection pool.

---

## 3. Verified Scenarios (Sample Data)

| Migration Type | Intended Level | Engine Result | Key Indicator |
| :--- | :--- | :--- | :--- |
| Concurrently Create Index | Safe | **Safe (30%)** | `LockLevel=4`, low impact. |
| Add Not Valid Constraint | Safe/Warning | **Safe (30%)** | Metadata-only change detected. |
| Alter Column Type | Danger | **Danger (85%)** | `RewriteRequired=true`, High `T_ddl`. |
| Set Not Null (Heavy) | Danger | **Danger (85%)** | Full table scan required during blocking. |
