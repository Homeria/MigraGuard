# ⚙️ MigraGuard v3.2 Implementation Details

This document provides a technical deep-dive into the core modules of MigraGuard, structured according to the Service Oriented Architecture (SOA) implemented in v3.2.

---

## 1. Core SQL Parser (`internal/parser/`)

- **Objective:** Parses DDL statements to identify target tables, lock levels, and whether a table rewrite is required ($F_{rewrite}$).
- **Library:** `github.com/pganalyze/pg_query_go/v5`.
- **$F_{rewrite}$ Detection:** Analyzes `ALTER TABLE` subtypes. Operations like `AT_AlterColumnType`, `AT_SetNotNull`, and `AT_AddColumn` with non-volatile defaults are identified as "Full Rewrite" operations.
- **Lock Mapping:** Dynamically maps DDL statements to PostgreSQL lock levels, accounting for `CONCURRENTLY` modifiers for index operations.

## 2. Background Collector (`internal/db/collector.go`)

- **Objective:** Orchestrates 24/7 metric collection and persistence with interval-based precision (Delta Collection).
- **Delta Computation Logic (v3.3):**
  - **Fetch:** Retrieves cumulative query stats from `pg_stat_statements` (`Current`).
  - **Lookup:** Retrieves the last known raw stats from the `original_pg_stat_statements` table (`Previous`).
  - **Compute Delta:** $\Delta = Current - Previous$. If $Current < Previous$ (indicating a DB reset), $\Delta = Current$.
  - **Update:** Saves the `Current` raw data to `original_pg_stat_statements` (UPSERT) for the next cycle.
  - **Persist:** Saves the computed $\Delta$ to the `workload_snapshots` table.
- **Retention Management:** Periodically executes `PurgeOldSnapshots()` to delete data exceeding the threshold (default: 7 days) and optimizes physical storage using `VACUUM`.

## 3. Database Adapters (`internal/db/`)

MigraGuard utilizes an interface-driven adapter pattern for high extensibility.

- **Postgres Adapter:**
  - Responsible for retrieving live metrics from the production database.
  - Implements `ValidateSchema()` to ensure the input SQL targets existing database objects.
- **SQLite Adapter:**
  - **Time-Series Storage:** Stores both interval-based deltas (`workload_snapshots`) and the latest raw state (`original_pg_stat_statements`).
  - **Persistence Strategy:** Uses a persistent SQLite table for raw stats instead of memory, ensuring data continuity even after agent restarts.
  - **Metric Calculation:** $\lambda = \frac{\Delta Calls}{\Delta Time}$.
  - **Baseline Computation:** Calculates `AvgTPS1h` and `PeakTPS24h` from the delta-based time-series data.

## 4. Risk Evaluation Engine (`internal/engine/risk.go`)

- **The 5-Step Model Logic:**
  - **Step 1 ($T_{ddl}$):** Static classification of DDL risk (SAFE/WARNING/DANGER).
  - **Step 2 ($T_{block}$):** Estimates lock duration: $T_{block} \propto N_{rows} \times F_{rewrite} + ReplicationLag$.
  - **Step 3 ($C_{peak}$):** Predicts peak load impact: $C_{peak} = C_{active} + (\lambda_{final} \times T_{block})$.
  - **Step 4 ($T_{rec}$):** Forecasts rollback time using table size and dead tuple ratios.
  - **Step 5 (Risk Score):** Derives a percentage based on $C_{peak}$ vs. configured maximum capacity.
- **Weighted Lambda ($\lambda_{final}$):**
  - $\lambda_{final} = \max(\lambda_{curr}, \lambda_{avg\_1h} \times 1.2, \lambda_{peak\_24h} \times 0.8)$.

## 5. Domain Service Layer (`internal/service/`)

- **AgentService:** Decouples the lifecycle management of the collector from the CLI entry points. Handles graceful shutdown and ticker management.
- **AnalyzeService:** Orchestrates the multi-step analysis workflow:
  1. SQL Load & AST Parsing.
  2. Database Schema Validation.
  3. Risk Engine Execution.
  4. Response aggregation for the Reporter.

## 6. Multi-Format Reporter (`internal/reporter/`)

- **Console Reporter:** Provides color-coded, human-readable terminal output for immediate developer feedback.
- **Markdown Reporter:** Generates structured reports suitable for GitHub Pull Request comments or documentation.
- **Gatekeeping:** If any analyzed table results in a `Danger` level, the service layer signals a high-risk failure, leading to an `Exit Code 1` to block CI/CD pipelines.

---

## Technical Summary: The Data Journey
1. **Agent** pulls metrics from **Postgres** and caches them in **SQLite**.
2. **Analyze** parses the **SQL file** to determine "what is being changed".
3. **Analyze** cross-references the change with **SQLite's** historical traffic to determine "is it safe to change now".
4. If risk exceeds the threshold, the process terminates with an error code to prevent production incidents.
