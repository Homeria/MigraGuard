# 🖥️ Native: 05. Check (Health Check)

The `check` command quickly verifies that MigraGuard can initialize the PostgreSQL connection and SQLite storage it needs.

---

## 1. Execution Examples

**Bash**
```bash
./migraguard check --db "postgres://user:pass@host:5432/db"
```

**PowerShell**
```powershell
.\migraguard.exe check --db "postgres://user:pass@host:5432/db"
```

**CMD**
```cmd
migraguard.exe check --db "postgres://user:pass@host:5432/db"
```

---

## 2. Detailed Flag Reference

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--db` | - | (From Config) | **Target PostgreSQL URL**. Verifies that MigraGuard can connect to PostgreSQL and ping it. |
| `--sqlite` | - | (From Config) | **Target SQLite Path**. Verifies that MigraGuard can open the SQLite store and initialize its schema. |

---

## 3. Current Check Scope
1.  **PostgreSQL connection**: DSN parsing, connection pool creation, and ping.
2.  **SQLite initialization**: Opens the configured SQLite path and creates the base MigraGuard tables if needed.

Note: the current `check` command does not separately verify that `pg_stat_statements` is loaded, that `shared_preload_libraries` is configured, or that disk space is sufficient. Verify those items during database setup or when running the `agent`.
