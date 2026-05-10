# 🐳 Docker: 04. Export Sandbox

Export metrics from an existing SQLite sandbox database to a CSV file.

---

## 1. Execution (Unified)
The command is identical across all shells.
```bash
docker compose run --rm analyze export-sandbox \
  --input ./code/experiments/data/01_steady_normal.db \
  --output ./code/output.csv
```

---

## 2. Parameter Reference
- `--input`: Source SQLite file (starts with `./code/`).
- `--output`: Target CSV path (starts with `./code/`).
