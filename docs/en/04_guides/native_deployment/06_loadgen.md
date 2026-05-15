# 🖥️ Native: 06. LoadGen (Traffic Generator)

LoadGen is a standalone binary that stresses your PostgreSQL database with realistic fintech transactions to test monitoring and risk analysis under pressure.

---

## 1. Execution Examples

**Bash**
```bash
./loadgen --db "postgres://user:pass@host:5432/db" --conns 20 --profile steady
```

**PowerShell**
```powershell
.\loadgen.exe --db "postgres://user:pass@host:5432/db" --conns 20 --profile steady
```

**CMD**
```cmd
loadgen.exe --db "postgres://user:pass@host:5432/db" --conns 20 --profile steady
```

---

## 2. Detailed Flag Reference

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--db` | - | **(Required)** | **Target PostgreSQL URL**. The DSN for the database where load will be injected. |
| `--conns` | - | `10` | **Concurrency (Workers)**. Number of simultaneous threads executing transactions. Higher values increase lock contention. |
| `--profile` | - | `steady` | **Load Profile**. Defines the transaction mix: <br> - `steady`: Balanced 70/30 Read/Write. <br> - `flash-sale`: Intense 30/70 Read/Write. <br> - `read-heavy`: 95/5 Read-only browsing. |

---

## 3. Transaction Details
*   **Browse**: `SELECT` users, products, and inventory.
*   **Order**: A full transaction involving `inventory_stocks` update, `account_balances` deduction, and `orders` insertion.
