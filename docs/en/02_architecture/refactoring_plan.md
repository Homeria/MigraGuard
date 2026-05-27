# 🛡️ MigraGuard Architecture Refactoring Plan

This document details the proactive architectural refactoring plan established ahead of high-level feature integration such as `feat/adaptive-recommendation` (machine-learning/statistics-based threshold auto-recommender).

This refactoring will be safely conducted on a **dedicated separate refactoring branch** after the current `feature/simulation-deepening` (simulation depth scaling) branch is merged into `develop`.

---

## 🎯 Core Objectives
1. **Single Responsibility Principle (SRP):** Reduce coupling so that each source file possesses only one dedicated role and responsibility.
2. **Isolate Mathematics and Infrastructure:** Separate raw statistical wave calculations from database disk I/O layers to guarantee unit test independence.
3. **Enhance Extensibility and Readability:** Deconstruct the 5-step risk engine process into modularized individual files.

---

## 📂 1. Deconstruct the Core Data Models (`pkg/migraguard/types/models.go`)
Currently, data schemas for completely different domains are centralized in a single file, which poses a serious scalability bottleneck as new reporting features are added.

### [Current State and Issues]
- Time-series database row metrics, DDL static analysis results, 24-hour forecast timelines, and YAML schema models for sandbox generation are all declared in a single file of approximately 180 lines.

### [Refactoring Blueprint]
Deconstruct the models into 4 distinct, purpose-driven files within `pkg/migraguard/types/`:
- **`models_core.go`:** `WorkloadSnapshot`, `TableDynamicMetrics`, `BaselineStats` (pure DB state metrics and workload snapshots).
- **`models_analysis.go`:** `AnalysisResult`, `RiskConstants`, `RiskAnalysisReport`, `AnalysisResponse` (models utilized by the 5-step quantitative evaluation process).
- **`models_forecast.go`:** `ForecastTimeSlot`, `ForecastReport` (models for 24-hour virtual timelines and peak forecast evaluations).
- **`models_simulation.go`:** `SimulationScenario`, `VirtualPGState`, `SQLiteHistoryProfile`, `TimelineEvent` (YAML deserialization configuration schemas for sandbox generation).

---

## 📂 2. Isolate Sandbox Seeding and Profiling Logic (`pkg/migraguard/internal/infra/sqlite/sqlite_sandbox.go`)
Mathematical wave algorithms and disk write operations (SQL transactions) are heavily intertwined in a single database driver file.

### [Current State and Issues]
- The statistical equations handling peak shifts, day-of-week scale adjustments, double sine cycles, monotonic time warping, and raw noise calculations (`DefaultWorkloadProfiler`) are in the same file as `SandboxEngine` and `SeedScenario` which open SQLite file paths and execute transaction bulk inserts.
- Testing the profiling algorithms independently from SQL disk I/O is highly difficult.

### [Refactoring Blueprint]
Isolate the pure math engine from the physical storage controller:
- **`sandbox_profiler.go`:** Hosts the `WorkloadProfiler` interface and `DefaultWorkloadProfiler` implementations. Keeps raw statistics and wave warped computations strictly isolated. (This serves as a highly scalable foundation for adding adaptive parameter fitting algorithms later)
- **`sqlite_sandbox.go`:** `SandboxEngine` will receive a dependency-injected (DI) `WorkloadProfiler` interface and focus solely on SQLite transactions, bulk seeding, and parameter initialization.

---

## 📂 3. Decompose the 5-Step Risk Evaluator Strategies (`pkg/migraguard/internal/analyzer/risk_evaluator.go`)
Five sequential risk assessment engines are written serially inside a single source file.

### [Current State and Issues]
- Strategies satisfying the `StepEvaluator` interface (`DDLTimeEvaluator`, `BlockingTimeEvaluator`, `PeakConnectionEvaluator`, `RecoveryTimeEvaluator`, `RiskScoreEvaluator`) are listed one after another. 
- While they are currently composed of simple mathematical formulas, adding queueing theory refinements, replication topology complexities, and multi-resource bottlenecks will cause this file's code length to surge rapidly.

### [Refactoring Blueprint]
Decompose individual strategies into atomized files within `pkg/migraguard/internal/analyzer/`:
- **`risk_evaluator.go`:** Defines only the core `StepEvaluator` interface and orchestrator signatures.
- **`evaluator_01_ddl_time.go`:** Predicts table rewrite costs and static metadata execution times ($T_{ddl}$).
- **`evaluator_02_blocking.go`:** Computes active live lock block times by coupling concurrent weight levels ($T_{block}$).
- **`evaluator_03_connections.go`:** Evaluates peak connection spikes via arrival rate and block duration ($C_{peak}$).
- **`evaluator_04_recovery.go`:** Calculates backlog evacuation durations and checks permanent queue collapse ($T_{rec}$).
- **`evaluator_05_scoring.go`:** Normalizes scores against resource limits and determines danger/warning thresholds.

---

## 🚀 Anticipated Synergy
- **Clean Separation of Concerns (SoC):** Developers can safely tune or replace the statistical time-warping equations without worrying about sqlite driver connection leakage.
- **Prerequisite for Adaptive Recommendations:** Enables clean integration of statistical parameter estimation modules inside `sandbox_profiler.go` and the analysis schemas inside `models_analysis.go` without altering database schemas or core orchestration steps.
