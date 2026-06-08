# 🖥️ Native: 01. Agent (Collector)

The Agent is a background metric collector that polls your PostgreSQL database to build a historical workload profile.

---

## 1. Execution Examples

**Bash (Unix/macOS/Linux)**
```bash
./migraguard agent --db "postgres://user:pass@localhost:5432/db"
```

**PowerShell (Windows)**
```powershell
.\migraguard.exe agent --db "postgres://user:pass@localhost:5432/db"
```

**CMD (Windows)**
```cmd
migraguard.exe agent --db "postgres://user:pass@localhost:5432/db"
```

---

## 2. Detailed Flag Reference

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--db` | - | (From Config) | **Target PostgreSQL Connection String (DSN)**. Specifies the database to monitor. Must have access to `pg_stat_statements`. |
| `--sqlite` | - | `migraguard.db` | **Local Storage Path**. The SQLite file where workload history is persisted. This file is used later by the `analyze` command. |
| `--interval` | - | `60` | **Polling Interval (Seconds)**. Defines how frequently the agent captures metrics. A lower value provides higher precision but increases monitoring overhead. |
| `--retention` | - | `7` | **Data Retention (Days)**. The agent will automatically prune metrics older than this value to manage disk space. |
| `--tables` | - | - | **Table-level dynamic metric targets**. A comma-separated list of tables (for example, `orders,users`). If omitted, the agent still records `pg_stat_statements` workload snapshots, but it does not write per-table size/TPS/P99/connection metrics into `table_metrics`. |

---

## 3. Best Practices
*   **Production Use**: Run the agent as a system service (e.g., `systemd` or Windows Service) to ensure continuous data collection.
*   **Precision**: For high-traffic systems, an interval of `30` is recommended for more granular risk assessment.
*   **Analysis targets**: If you plan to use `analyze --forecast` or table-level baseline analysis, pass the tables you care about with `--tables`.
