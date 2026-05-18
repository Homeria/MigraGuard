# Level 3: Implementation Traceability Matrix

This document maps the architectural domains and workflows described in Level 0-2 to the actual Go source code and functions.

---

## 1. Command Domain (Interface Layer)
*Entry points for user interactions via Cobra CLI.*

| Workflow | Entry Point (Go File) | Key Function/Method | Description |
| :--- | :--- | :--- | :--- |
| `analyze` | `cmd/migraguard/analyze.go` | `runAnalyze()` | Primary orchestrator for the analysis command. |
| `simulate` | `cmd/migraguard/simulate.go` | `runSimulate()` | Handles scenario loading and sandbox initialization. |
| `check` (Agent) | `cmd/migraguard/check.go` | `runCheck()` | Starts the background metric collection agent. |

---

## 2. Application Domain (Service Layer)
*Orchestration logic that bridges the CLI and the Core SDK.*

| Service | Source File | Core Logic Method | Role |
| :--- | :--- | :--- | :--- |
| **Analyze Service** | `pkg/migraguard/internal/app/analyze_service.go` | `Run()` | Coordinates SQL parsing, risk evaluation, and forecast CSV export. |
| **Simulate Service** | `pkg/migraguard/internal/app/simulate_service.go` | `Run()` | Manages scenario YAML unmarshaling and sandbox seeding. |
| **Agent Service** | `pkg/migraguard/internal/app/agent_service.go` | `Run()` | Implements the infinite loop for periodic metric harvesting. |

---

## 3. Domain Logic: Risk Engine (Strategy Layer)
*The mathematical core of the system, implementing the 5-step model.*

| Architectural Component | Source File | Key Symbol/Function | Description |
| :--- | :--- | :--- | :--- |
| **Orchestrator** | `pkg/migraguard/internal/analyzer/risk_calculator.go` | `AnalyzeRisk()` | Executes the sequence of StepEvaluators for a single point-in-time. |
| **Predictive Engine** | `pkg/migraguard/internal/analyzer/risk_calculator.go` | `AnalyzeForecast()` | Runs 24-hour virtual simulations with the **TPS Tie-Breaker** logic. |
| **SQL Parser** | `pkg/migraguard/internal/analyzer/sql_parser.go` | `ParseSQL()` | Utilizes `pg_query_go` to convert SQL to AST and identify DDL types. |
| **Evaluators (Steps 1-5)** | `pkg/migraguard/internal/analyzer/risk_evaluator.go` | `StepEvaluator` (Interface) | Strategy pattern implementation for $T_{ddl}, T_{block}, C_{peak}, T_{rec}$, and final Score. |

---

## 4. Infrastructure Domain (Adapter Layer)
*Interfaces for interacting with databases.*

| Adapter | Source File | Key Responsibility | Implementation Details |
| :--- | :--- | :--- | :--- |
| **PostgreSQL Adapter** | `pkg/migraguard/internal/infra/postgres/adapter.go` | `FetchTableDynamicMetrics()` | Real-time metric extraction from `pg_stat_statements` and system catalogs. |
| **SQLite Repository** | `pkg/migraguard/internal/infra/sqlite/sqlite_repository.go` | `InitializeSchema()` | Management of the local metrics storage and sandbox files. |
| **Virtual PG Adapter** | `pkg/migraguard/internal/infra/sqlite/sqlite_virtual.go` | `FetchTableDynamicMetrics()` | **Simulation Core**: Mocks PostgreSQL behavior by reading from SQLite sandbox data. |
| **Sandbox Engine** | `pkg/migraguard/internal/infra/sqlite/sqlite_sandbox.go` | `SeedScenario()` | Generates time-series data based on Sine waves and Noise variance. |

---

## 5. Visualization Pipeline
*Bridge between Go analysis and Python reporting.*

| Component | Source File | Trigger Point | Description |
| :--- | :--- | :--- | :--- |
| **CSV Exporter** | `pkg/migraguard/internal/app/analyze_service.go` | `exportForecastCSV()` | Generates `predictive_forecast.csv` after successful analysis. |
| **Python Plotter** | `tools/visualization/analyze/plot_predictive_heatmap.py` | `plot_predictive_heatmap()` | Matplotlib implementation of the **Dual Y-Axis Heatmap** (TPS & P99). |
| **Orchestrator Script**| `scripts/cmd/analyze/run-predictive-forecast.bat` | N/A | Batch/Shell script that chains Go CLI and Python Plotter execution. |
