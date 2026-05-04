# 🧪 Sandbox File-based Simulation Architecture

This document describes the design and implementation of the **Simulation Sandbox** in MigraGuard v3.8, optimized for academic research and statistical validation.

---

## 1. Overview

The Simulation Sandbox allows researchers to evaluate migration risks in a controlled environment without a live PostgreSQL instance. By defining scenarios in YAML, MigraGuard generates independent SQLite sandbox files and performs risk analysis using virtualized database states.

### Core Objectives
- **Reproducibility**: Ensure identical results for the same scenario YAML and SQL.
- **Isolation**: Prevent interference with production metrics stored in `migraguard.db`.
- **Portability**: Generate physical `.db` files that can be audited using standard SQLite tools.
- **Statistical Analysis**: Export raw risk metrics to CSV for external processing (Python, Excel, etc.).

---

## 2. Declarative Scenario (YAML)

Experiments are defined using a declarative syntax. This enables version control of experimental setups.

```yaml
# Example: experiments/spike_test.yaml
experiment_name: "High Traffic Spike"
description: "Simulating a 5x traffic increase during a flash sale."
target_table: "orders"

# Virtual PostgreSQL State
pg_state:
  table_size_mb: 2048
  active_connections: 350
  p99_time_ms: 45.0
  current_tps: 3000.0

# SQLite History Generation Rules
sqlite_history:
  pattern: "spike" # patterns: steady, sine, spike, random
  days: 7
  base_tps: 500.0
  peak_tps: 4500.0
```

---

## 3. Implementation Phases

### Phase 1: Simulation Sandbox (`feat/simulation-sandbox`)
- Implement YAML parser for scenario definitions.
- Create a sandbox engine that generates a fresh `{experiment_name}.db` for each run.
- Implement mathematical data generators (Sine, Spike, etc.) to populate SQLite tables.

### Phase 2: Virtual PG Adapter (`feat/virtual-pg-adapter`)
- Create `VirtualPGAdapter` implementing the `PostgresClient` interface.
- Map YAML `pg_state` values to adapter responses (Table size, Conns, P99).
- Enable `mg analyze --mock` to switch the SDK to use the virtual adapter.

### Phase 3: Research Metrics Logger (`feat/research-csv-export`)
- Implement a structured logger that captures all 5-step risk metrics ($T_{ddl}, T_{block}, C_{peak}, T_{rec}$, Score).
- Add CSV export functionality to the CLI.

---

## 4. Usage Workflow

1. **Design**: Write a scenario YAML.
2. **Execute**: Run `mg simulate --scenario path/to/yaml --sql path/to/sql`.
3. **Verify**: Inspect the generated `.db` file using `sqlite3`.
4. **Analyze**: Use the generated `results.csv` for research papers or presentations.
