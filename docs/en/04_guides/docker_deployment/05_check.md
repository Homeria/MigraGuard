# 🐳 Docker: 05. Check (Health Check)

Verify connectivity and status.

---

## 1. Execution (Unified)
The command is identical across all shells.
```bash
docker compose run --rm analyze check
```
---
- Checks `db` container connectivity.
- Checks shared `migraguard.db` write access.
