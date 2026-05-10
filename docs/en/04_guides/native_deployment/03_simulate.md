# 🖥️ Native: 03. Simulate (Scenario Generation)

The Simulate command creates virtual time-series metrics based on mathematical models for research and testing.

---

## 1. Execution Examples

**Bash**
```bash
./migraguard simulate --scenario experiments/scenarios/03_spike_flash_sale.yaml
```

**PowerShell**
```powershell
.\migraguard.exe simulate --scenario .\experiments\scenarios\03_spike_flash_sale.yaml
```

**CMD**
```cmd
migraguard.exe simulate --scenario experiments\scenarios\03_spike_flash_sale.yaml
```

---

## 2. Detailed Flag Reference

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--scenario` | `-s` | **(Required)** | **Input Scenario Path**. Points to the YAML file defining the workload profile (TPS, Events, Spikes). |
| `--force` | `-f` | `false` | **Force Overwrite**. If the target sandbox database already exists, this flag allows the tool to overwrite it. |
| `--csv` | - | - | **CSV Export Path**. If specified, the tool will also export the generated metrics to this CSV file. Use `-` for stdout. |
| `--no-db` | - | `false` | **Ephemeral Mode**. If used with `--csv`, the generated SQLite file will be deleted after the CSV is exported. |

---

## 3. CSV Export Redirection

**Bash / CMD**
```bash
./migraguard simulate -s scenario.yaml --csv - > metrics.csv
```

**PowerShell**
```powershell
.\migraguard.exe simulate -s scenario.yaml --csv - | Out-File -FilePath metrics.csv -Encoding utf8
```
