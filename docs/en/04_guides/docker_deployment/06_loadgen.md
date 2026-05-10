# 🐳 Docker: 06. LoadGen (Traffic Generator)

Simulate real-world fintech traffic on the `db` container.

---

## 1. Execution (Unified)
The command is identical across all shells.
```bash
docker compose run -d --name stress-test load-generator --conns 50 --profile flash-sale
```

---

## 2. Flag Reference
| Flag | Default | Description |
| :--- | :--- | :--- |
| `--conns` | `10` | Worker concurrency. |
| `--profile` | `steady` | `steady`, `flash-sale`, `read-heavy`. |
