# 🐳 Docker: 01. Agent (Collector)

The Agent service runs in the background to harvest workload patterns.

---

## 1. Execution (Unified)
The command is identical across all shells (Bash, PowerShell, CMD).
```bash
docker compose logs -f agent
```

## 2. Configuration
To change agent settings, modify the `command` in `docker-compose.yml` and restart:
```bash
docker compose up -d agent
```

---

## 3. Flag Reference
| Flag | Default | Description |
| :--- | :--- | :--- |
| `--db` | (Config) | Target DB. Use service name `db`. |
| `--sqlite` | `/app/data/migraguard.db` | Internal volume path. |
| `--interval` | `60` | Polling frequency (sec). |
