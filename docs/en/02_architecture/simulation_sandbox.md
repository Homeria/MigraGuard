# 🧪 Sandbox File-based Simulation Architecture

This document describes the design and implementation of the **Simulation Sandbox** in MigraGuard v3.9, optimized for academic research and statistical validation.

---

## 1. Overview

The Simulation Sandbox allows researchers to evaluate migration risks in a controlled environment without a live PostgreSQL instance. By defining scenarios in YAML, MigraGuard generates independent SQLite sandbox files and performs risk analysis using virtualized database states.

In the v3.9 release, the mathematical wave generator (Profiler) has been completely decoupled, and robust error guards such as thread-safe local random sources (`rng`) and `tx.Prepare` error checkers are integrated to maximize mock database seeding integrity.

### Core Objectives
- **Reproducibility**: Ensure identical results for the same scenario YAML and SQL configurations.
- **Isolation**: Prevent interference with production metrics stored in `migraguard.db`.
- **Portability**: Generate physical `.db` files that can be audited using standard SQLite tools.
- **Statistical Analysis & Visualization**: Expose structured results to plot 24h individual wave heatmaps and master analysis CSV/PNG chart reports.

---

## 2. v3.9 Declarative Realistic Scenario (YAML)

v3.9 scenarios embed advanced physical workload attributes such as asymmetric time warping, peak shift offsets, weekend dormancy, and daily amplitude scaling parameters.

```yaml
# Example: experiments/scenarios/04_sine_daily_cycle.yaml
experiment_name: "sine_daily_cycle"
description: "Realistic daily traffic swing simulation with high skewness"
target_table: "orders"

# Virtual PostgreSQL Metrics State (Now)
pg_state:
  table_size_mb: 2048
  replication_lag_s: 0
  active_connections: 50
  p99_time_ms: 10.0
  current_tps: 150.0

# v3.9 Advanced SQLite History Generation Rules
sqlite_history:
  days: 7
  interval_minutes: 30
  base_tps: 50.0
  peak_tps: 500.0
  noise_variance: 0.10
  asymmetric_skew: 0.25      # Time-warping skewness factor
  peak_shift_hours: 1.5      # Peak hour fluctuation offset
  weekly_pattern: true       # 40% Traffic reduction during weekends
  events:                    # Sudden traffic surge events
    - name: "evening-burst"
      start_day: 3
      start_hour: 18
      duration_h: 4
      multiplier: 1.8
```

---

## 3. Implementation Phases & Architecture Decoupling

### Phase 1: Pure Statistical Model (Profiler) Decoupling
- `DefaultWorkloadProfiler` is decoupled into `sandbox_profiler.go` to handle pure wave equations (Sine, Time-warped Skew, Noise) independently.
- Global math/rand resource locks are eliminated by introducing a thread-safe random number generator `rand.New(rand.NewSource(now.UnixNano()))` for massive concurrent seeding.

### Phase 2: Virtual PG Adapter (`VirtualPGAdapter`)
- Created `VirtualPGAdapter` (`sqlite_virtual.go`) satisfying the `PostgresClient` interface.
- Mapped YAML `pg_state` values to virtual adapter responses (Table size, Conns, P99, TPS).
- Enforced database connection exception safety via robust `rows.Err()` checks, securing row scans from silent omissions.

### Phase 3: Research Metrics & L2/L3 Orchestration
- Dynamic multi-dimensional mock seedings are packaged under a single timestamped directory (`experiments/reports/batch_runs/[TIMESTAMP]/`), automatically plotting 24h individual heatmaps and comprehensive research charts.

---

## 4. CLI Options & Usage Workflow

### 1) Seed a Single Simulation Sandbox
Run the sandbox simulation command to create a virtual SQLite database according to the scenario configuration:
```bash
./build/migraguard simulate --scenario experiments/scenarios/04_sine_daily_cycle.yaml --force
```

### 2) L2 Single Scenario Pipeline Orchestration
Trigger a single scenario simulation, run dynamic static/dynamic risk calculations, and generate individual heatmaps:
```bash
bash scripts/sh/run_scenario_pipeline.sh experiments/scenarios/04_sine_daily_cycle.yaml
```

### 3) L3 Global Benchmark Pipeline Orchestration
Evaluate all 20 realistic scenarios concurrently with a single command, producing complete research results and master chart reports:
```bash
bash scripts/sh/run_global_pipeline.sh
```
