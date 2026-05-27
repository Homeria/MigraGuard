# 🐳 Docker: 03. Simulate (Scenario Generation)

Generate deterministic workload scenarios.

---

## 1. Execution (Unified)
This command is identical across all shells.
```bash
docker compose run --rm analyze simulate --scenario ./code/experiments/scenarios/03_spike_flash_sale.yaml
```

## 2. Direct CSV Export to Host
If using redirection to save raw metrics, note the PowerShell difference.

**Bash / CMD**
```bash
docker compose run --rm analyze simulate -s ./code/scen.yaml --csv - > host_metrics.csv
```

**PowerShell**
```powershell
docker compose run --rm analyze simulate -s ./code/scen.yaml --csv - | Out-File -FilePath host_metrics.csv -Encoding utf8
```

---

## 3. Flag Reference
| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--scenario` | `-s` | **(Required)** | Input YAML path (starts with `./code/`). |
| `--force` | `-f` | `false` | Force overwrite the sandbox database if it already exists. |
| `--csv` | - | - | Export path. Use `-` for stdout redirection. |
| `--no-db` | - | `false` | Ephemeral mode. When used with `--csv`, removes the SQLite file after export. |
