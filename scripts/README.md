# 🛡️ MigraGuard Ultra-Descriptive Automation Scripts

Scripts are organized by platform and task type with highly descriptive naming conventions.

## 📂 Hierarchical Structure
`scripts/{platform}/{task_type}/{descriptive-name}.ext`

- **Platforms**: `ps1` (PowerShell - Recommended), `cmd` (Batch), `sh` (Bash)
- **Task Types**: 
  - `sandbox`: Simulation data generation (Scenario -> DB/CSV).
  - `analyze`: Risk engine execution (Data -> Report).

---

## 🏗️ 1. Sandbox Scripts (Data Generation)
Located in `scripts/{platform}/sandbox/`

| Descriptive Script Name | Purpose | Output |
| :--- | :--- | :--- |
| `gen-db-from-scenario` | Create 1 sandbox DB from YAML | `experiments/data/*.db` |
| `gen-db-from-all-scenarios` | Create all sandbox DBs in folder | `experiments/data/*.db` |
| `gen-csv-from-scenario` | Export 1 metrics CSV from YAML | `Specified Path` |
| `gen-csv-from-all-scenarios` | Export all metrics CSVs to reports | `experiments/reports/metrics/*.csv` |

---

## 🔬 2. Analyze Scripts (Risk Assessment)
Located in `scripts/{platform}/analyze/`

| Descriptive Script Name | Purpose | Output Formats |
| :--- | :--- | :--- |
| `run-analysis-from-sandbox` | Analyze 1 SQL against 1 Sandbox DB | Console, CSV, Markdown, TXT |
| `run-analysis-from-live` | Analyze 1 SQL against Live Postgres | Console, CSV, Markdown, TXT |
| `run-batch-research-report` | Full cross-product batch analysis | `experiments/reports/research_results.csv` |

---

### 💡 Advanced Usage
- **TXT/CSV Export**: Analysis scripts support an optional `[output_type]` (csv, markdown, console). To save to a text file, use the `[report_path]` parameter in PS1 or standard redirection `> result.txt` in CMD/Bash.
- **Master Report**: `run-batch-research-report` is the main tool for collecting academic statistics for thesis or presentations.
