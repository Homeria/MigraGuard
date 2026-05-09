# 🧩 MigraGuard SQL Parser Logic Specification

This document describes the technical mechanisms used by MigraGuard to analyze DDL SQL, determine lock levels, and detect potential table rewrites.

---

## 1. Overview

MigraGuard utilizes `pg_query_go`, which wraps the official PostgreSQL C-parser. This ensures 100% syntax compatibility and accuracy for production SQL analysis.

---

## 2. Analysis Pipeline

1.  **AST Generation**: Converts the SQL string into an Abstract Syntax Tree.
2.  **DDL Identification**: Extracts primary DDL nodes such as `AlterTableStmt`, `CreateStmt`, and `IndexStmt`.
3.  **Risk Indicator Detection**:
    *   **Lock Level**: Assigns a level (1-8) based on the internal PostgreSQL lock matrix.
    *   **Rewrite Required**: Detects operations requiring full table scans and storage rewrites (e.g., column type changes).
    *   **Metadata Only**: Identifies lightweight operations that only update system catalogs.

---

## 3. Example Classification Logic

| SQL Statement | Classification | Reason |
| :--- | :--- | :--- |
| `ALTER TABLE orders ADD COLUMN age int;` | **Safe (Metadata)** | Adding a column only updates system metadata. |
| `ALTER TABLE orders ALTER COLUMN no TYPE bigint;` | **Danger (Rewrite)** | Type changes require rewriting all existing data rows. |
| `CREATE INDEX idx_name ON users(name);` | **Warning (Lock)** | Standard index creation holds a `ShareLock`, blocking writes. |

---

## 4. Architectural Advantages
- **Reliability**: Uses the official PG parser for maximum syntax detection accuracy.
- **Predictability**: Establishes the lock characteristics and execution baseline during the static analysis phase.
