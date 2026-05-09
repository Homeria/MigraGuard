# 📜 MigraGuard Script Manual

This document provides detailed instructions for MigraGuard's automation scripts in both KR and EN.

---

## 📂 1. Sandbox Group (Data Generation)
Location: `scripts/{platform}/sandbox/`

### 1.1. gen-db-from-scenario
- **Role**: Generates an SQLite sandbox from a single YAML scenario.
- **PowerShell**: `.\scripts\ps1\sandbox\gen-db-from-scenario.ps1 -Scenario 01_steady.yaml`

### 1.2. gen-db-from-all-scenarios
- **Role**: Batch generates all scenarios.
- **Bash**: `./scripts/sh/sandbox/gen-db-from-all-scenarios.sh`

---

## 🔬 2. Analyze Group (Risk Analysis)
Location: `scripts/{platform}/analyze/`

### 2.1. run-analysis-from-sandbox
- **Role**: Performs offline analysis using generated sandbox files.
- **CMD**: `scripts\cmd\analyze\run-analysis-from-sandbox.bat steady.db add_col.sql`

### 2.2. run-batch-research-report
- **Role**: Executes exhaustive cross-analysis and generates research reports.
- **PowerShell**: `.\scripts\ps1\analyze\run-batch-research-report.ps1`

---

## 📊 3. Visualization Tools
- **Sandbox Metrics**: `python tools/visualization/sandbox/visualize_all_scenarios.py`
- **Analysis Results**: `python tools/visualization/analyze/visualize_all_reports.py`
