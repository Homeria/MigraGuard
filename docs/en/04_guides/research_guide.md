# 📊 MigraGuard Research Data & Analysis Guide

This guide describes how to extract large-scale experimental data using the **Batch Research** feature and generate statistical assets for academic reports using visualization tools.

---

## 1. Overview

MigraGuard serves as an experimental platform to quantitatively research the correlation between DDL operations and operational traffic.

*   **Exhaustive Analysis**: Cross-analyzes all scenarios and test DDLs to secure hundreds of precision data points simultaneously.
*   **Automated Visualization**: Generates load profiles and risk heatmaps via Python scripts.

---

## 2. Workflow

### Step 1: Research Dataset Setup
Generate SQLite sandbox files for all scenarios.
```powershell
.\scripts\ps1\sandbox\gen-db-from-all-scenarios.ps1
```

### Step 2: Batch Analysis & CSV Extraction
Execute cross-analysis combining the datasets with all cases in `experiments/ddl`.
```powershell
.\scripts\ps1\analyze\run-batch-research-report.ps1
```
*   **Result**: Generates `experiments\reports\research_results.csv`.

### Step 3: Data Visualization
Use the provided Python tools to generate graphs.

**A. Risk Heatmap Generation**
```bash
python tools/visualization/analyze/visualize_all_reports.py
```

**B. Load Profile Visualization**
```bash
python tools/visualization/sandbox/visualize_all_scenarios.py
```

---

## 3. Research Tips

### 3.1. Pivot Table Analysis
Analyze risk trends by scenario using `RiskScore`. Specifically, charting the relationship between `TableSize` and `T_ddl` can prove the system's predictive accuracy.

### 3.2. Advanced Python Analysis
Researchers can use Pandas to filter "Danger" cases and quantitatively analyze their relationship with infrastructure thresholds ($C_{max}$).
