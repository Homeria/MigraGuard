# 🐳 Docker: 02. Analyze (Risk assessment)

Run high-fidelity risk analysis. **Note the mandatory `./code/` prefix for host files.**

---

## 💡 Mandatory: Path Mapping
The container sees your host files under the `./code/` directory.
- **Fail**: `analyze experiments/ddl/mig.sql`
- **Success**: `analyze ./code/experiments/ddl/mig.sql`

---

## 1. Execution (Unified)
This command works in any shell.
```bash
docker compose run --rm analyze analyze ./code/experiments/ddl/001_safe_add_column_orders.sql
```

## 2. Saving Reports (Shell Specific)
Redirection syntax differs if you want to save the output to a file on your host.

**Bash / CMD**
```bash
docker compose run --rm analyze analyze ./code/mig.sql -o markdown > report.md
```

**PowerShell**
```powershell
docker compose run --rm analyze analyze ./code/mig.sql -o markdown | Out-File -FilePath report.md -Encoding utf8
```

---

## 3. Flag Reference
| Flag | Default | Description |
| :--- | :--- | :--- |
| `--sandbox` | - | Offline mode. Path must start with `./code/`. |
| `--output` | `console` | `console`, `markdown`, `csv`. |
