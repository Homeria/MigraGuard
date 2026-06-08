# 🛡️ MigraGuard Architecture Refactoring Plan & Results Report

This document details the proactive architectural refactoring plan established ahead of high-level feature integration such as `feat/adaptive-recommendation` (machine-learning/statistics-based threshold auto-recommender), and records the final successful execution results.

> [!NOTE]
> All refactoring items outlined in this architectural document have been **100% successfully completed and verified as of 2026-05-27**.

---

## 🎯 Core Objectives & Accomplishments
1. **Single Responsibility Principle (SRP):** Reduced coupling so that each source file possesses only one dedicated role and responsibility. **[COMPLETED]**
2. **Isolate Mathematics and Infrastructure:** Separated raw statistical wave calculations from database disk I/O layers to guarantee unit test independence. **[COMPLETED]**
3. **Enhance Extensibility and Readability:** Deconstructed the 5-step risk engine process into modularized individual files and dedicated packages. **[COMPLETED]**
4. **Decoupled Logging and Exception Safety:** Eliminated silent error suppression, enforced database rows loop error checks, and initialized thread-local RNGs. **[NEWLY ADDED & COMPLETED]**

---

## 📂 1. Deconstructed Core Data Models (`pkg/migraguard/types/`)
Domain models have been successfully split into 4 distinct, purpose-driven files to resolve scalability issues:
- **`models_core.go`:** `WorkloadSnapshot`, `TableDynamicMetrics`, `BaselineStats` (pure DB state metrics and workload snapshots).
- **`models_analysis.go`:** `AnalysisResult`, `RiskConstants`, `RiskAnalysisReport`, `AnalysisResponse` (models utilized by the 5-step quantitative evaluation process).
- **`models_forecast.go`:** `ForecastTimeSlot`, `ForecastReport` (models for 24-hour virtual timelines and peak forecast evaluations).
- **`models_simulation.go`:** `SimulationScenario`, `VirtualPGState`, `SQLiteHistoryProfile`, `TimelineEvent` (YAML deserialization configuration schemas for sandbox generation).

---

## 📂 2. Isolated Sandbox Seeding and Profiling Logic (`pkg/migraguard/internal/infra/sqlite/`)
Pure mathematical formulas are now completely isolated from the database access layer:
- **`sandbox_profiler.go`:** Hosts the `WorkloadProfiler` interface and `DefaultWorkloadProfiler` implementations. Keeps raw statistics and wave warped computations strictly isolated.
- **`sqlite_sandbox.go`:** `SandboxEngine` is injected with the `WorkloadProfiler` interface and focuses solely on SQLite transactions, bulk seeding, and parameter initialization.

---

## 📂 3. Decomposed 5-Step Risk Evaluator Strategies (`pkg/migraguard/internal/analyzer/`)
Sequential risk assessment engines have been separated into a dedicated subpackage (`evaluators`) with atomized individual files:
- **`evaluator.go`:** Defines only the core `StepEvaluator` interface and orchestrator signatures.
- **`ddl_time.go`:** Predicts table rewrite costs and static metadata execution times ($T_{ddl}$).
- **`blocking.go`:** Computes active live lock block times by coupling concurrent weight levels ($T_{block}$).
- **`connections.go`:** Evaluates peak connection spikes via arrival rate and block duration ($C_{peak}$).
- **`recovery.go`:** Calculates backlog evacuation durations and checks permanent queue collapse ($T_{rec}$).
- **`scoring.go`:** Normalizes scores against resource limits and determines danger/warning thresholds.

---

## 📂 4. [Newly Completed] Exception Safety & Structured Logging
To secure the v3.9 declarative simulation orchestrator, four high-priority refinement steps were successfully completed:
- **Structured Simulation Logging**: Replaced generic console standard print operations (`fmt`) inside `simulate_service.go` with structured logger (`types.Logger`) invocations, and established DI propagation in `client.go`.
- **Enforced Database `rows.Err()` Checks**: Enforced `rows.Err()` checks immediately after `rows.Next()` loop ends in `sqlite_analyzer.go`, `sqlite_virtual.go`, and `pg_client_impl.go` to capture hidden database driver faults, and resolved silent row omissions by replacing generic `continue` statements with explicit error propagation.
- **SQLite Error Propagation in Risk Engine**: Eliminated blank identifiers (`_`) for database queries inside `risk_calculator.go`, establishing strict error checking and custom wraps for statistical queries.
- **Sandbox Prepare & Decoupled Thread-Local RNG**: Implemented strict error checks for `tx.Prepare` statements inside `sqlite_sandbox.go` with safe `defer Close()` wrappers, and decoupled global `math/rand` contention by adopting a thread-local random source generator (`rand.Rand`).

---

## 🚀 Architectural Benefits
- **Clean Separation of Concerns (SoC):** Mathematical wave-warping formulas can be tuned or replaced safely without worrying about resource leaks or driver logic bugs.
- **Prerequisite for Adaptive Recommendations:** Enables clean integration of statistical parameter estimation modules inside `sandbox_profiler.go` and analysis schemas inside `models_analysis.go` without altering database schemas or core orchestration steps.
