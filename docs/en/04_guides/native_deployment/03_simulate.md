# 🖥️ Native: 03. Simulate (Scenario Generation)

The Simulate command creates virtual time-series metrics based on mathematical models for research and testing.

---

## 1. Execution Examples

**Bash**
```bash
./migraguard simulate --scenario experiments/scenarios/05_spike_flash_sale.yaml
```

**PowerShell**
```powershell
.\migraguard.exe simulate --scenario .\experiments\scenarios\05_spike_flash_sale.yaml
```

**CMD**
```cmd
migraguard.exe simulate --scenario experiments\scenarios\05_spike_flash_sale.yaml
```

`simulate` creates `<experiment_name>.db` in the current working directory, where `experiment_name` comes from the scenario YAML. The current CLI does not provide a flag to set the sandbox DB output path directly.

---

## 2. Detailed Flag Reference

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--scenario` | `-s` | **(Required)** | **Input Scenario Path**. Points to the YAML file defining the workload profile (TPS, Events, Spikes). |
| `--force` | `-f` | `false` | **Force Overwrite**. If the target sandbox database already exists, this flag allows the tool to overwrite it. |
| `--csv` | - | - | **CSV Export Path**. If specified, the tool will also export the generated metrics to this CSV file. Use `-` for stdout. |
| `--no-db` | - | `false` | **Ephemeral Mode**. If used with `--csv`, the generated SQLite file will be deleted after the CSV is exported. |

Note: `--no-db` only has an effect when `--csv` is also provided. If `--no-db` is used without `--csv`, the generated DB remains on disk.

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

Use only `--csv <path>` if you want to keep both the CSV and the generated DB. Use `--csv <path> --no-db` when you only want the CSV and do not want to keep the temporary DB.

---

## 4. Symmetrical Batch Orchestration Pipelines (Orchestration Scripts)

Using the Standalone L1 micro-modules and L2/L3 orchestrators built in v3.9, you can automate massive multi-capacity DDL batch simulations with a single command. It is symmetrically designed for all platforms (Linux Bash, Windows PowerShell, and Windows CMD Batch).

### A. Linux Bash Pipeline

- **L1 Standalone Modules**:
  - `bash scripts/sh/modules/simulate_scenario.sh <scenario_yaml>`
  - `bash scripts/sh/modules/dispatch_db.sh <seed_db> <case_yaml>`
  - `bash scripts/sh/modules/analyze_single_ddl.sh <ddl> <db> <config> <out_csv>`
- **L2 Single Scenario Pipeline**:
  ```bash
  bash scripts/sh/run_scenario_pipeline.sh experiments/scenarios/02_commuter_daily_rush.yaml
  ```
- **L3 Global Orchestrator**:
  ```bash
  bash scripts/sh/run_global_pipeline.sh
  ```

### B. Windows PowerShell Pipeline

- **L1 Standalone Modules**:
  - `.\scripts\ps1\modules\simulate_scenario.ps1 <scenario_yaml>`
  - `.\scripts\ps1\modules\dispatch_db.ps1 <seed_db> <case_yaml>`
  - `.\scripts\ps1\modules\analyze_single_ddl.ps1 <ddl> <db> <config> <out_csv>`
- **L2 Single Scenario Pipeline**:
  ```powershell
  .\scripts\ps1\run_scenario_pipeline.ps1 .\experiments\scenarios\02_commuter_daily_rush.yaml
  ```
- **L3 Global Orchestrator**:
  ```powershell
  .\scripts\ps1\run_global_pipeline.ps1
  ```

### C. Windows CMD Batch Pipeline

- **L1 Standalone Modules**:
  - `call .\scripts\cmd\modules\simulate_scenario.cmd <scenario_yaml>`
  - `call .\scripts\cmd\modules\dispatch_db.cmd <seed_db> <case_yaml>`
  - `call .\scripts\cmd\modules\analyze_single_ddl.cmd <ddl> <db> <config> <out_csv>`
- **L2 Single Scenario Pipeline**:
  ```cmd
  call .\scripts\cmd\run_scenario_pipeline.cmd experiments\scenarios\02_commuter_daily_rush.yaml
  ```
- **L3 Global Orchestrator**:
  ```cmd
  call .\scripts\cmd\run_global_pipeline.cmd
  ```

