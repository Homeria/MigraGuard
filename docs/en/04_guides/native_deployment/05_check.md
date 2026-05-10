# 🖥️ Native: 05. Check (Health Check)

The Check command performs a self-diagnostic on the MigraGuard environment and connectivity.

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
| `--db` | - | (From Config) | **Target PostgreSQL URL**. Verifies connection and checks if the `pg_stat_statements` extension is loaded. |
| `--sqlite` | - | (From Config) | **Target SQLite Path**. Verifies that the file exists and is writable by the current user. |

---

## 3. Diagnostic Checklist
1.  **PostgreSQL Connection**: Checks DNS resolution and authentication.
2.  **Extension Verification**: Confirms `pg_stat_statements` is in `shared_preload_libraries`.
3.  **SQLite Health**: Verifies disk space and write permissions.
