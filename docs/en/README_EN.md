# 🛡️ MigraGuard: PostgreSQL Migration Risk Gatekeeper

> **"Operate your DDL with Zero-Blindness."**
>
> MigraGuard is a high-fidelity risk prediction system that quantifies the impact of database migrations on live production traffic. It prevents service outages caused by invisible lock contentions by cross-analyzing DDL semantics with real-time workload patterns.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/Homeria/MigraGuard)](https://goreportcard.com/report/github.com/Homeria/MigraGuard)
[![Research Grade](https://img.shields.io/badge/Architecture-Research--Grade-blueviolet.svg)](#-mathematical-risk-model)

---

## 📖 Table of Contents
- [Why MigraGuard?](#-why-migraguard)
- [System Architecture](#-system-architecture)
- [Mathematical Risk Model](#-mathematical-risk-model)
- [Simulation Sandbox](#-simulation-sandbox)
- [Quick Start](#-quick-start)
- [Documentation Hub](#-documentation-hub)

---

## 🧐 Why MigraGuard?

In modern fintech and high-traffic systems, a single `ALTER TABLE` can paralyze the entire database due to **Exclusive Locks**. While traditional tools focus on syntax validation, MigraGuard focuses on **Operational Context**.

- **The Problem**: DDL execution in production is often a "leap of faith." Developers don't know how long it will take or how many connections will be blocked until it actually happens.
- **The Solution**: MigraGuard predicts the future by analyzing your SQL against your current TPS (Transactions Per Second), Latency (P99), and Replication Lag.

---

## 🏗️ System Architecture

MigraGuard follows an **SDK-First Modular Architecture**. For detailed visual maps and implementation logic, please refer to the refreshed documentation hierarchy:

- **[Level 0: System Context](./02_architecture/uml_diagrams/00_system_context.md)**: High-level overview of actors and components.
- **[Level 1: CLI Workflows](./02_architecture/uml_diagrams/01_cmd_workflows.md)**: Mid-level operational flows for primary commands.
- **[Level 1+: Implementation Flows](./02_architecture/uml_diagrams/01_cmd_implementation_flows.md)**: Granular CLI execution paths with file/function nodes.
- **[Level 2: Core Deep-Dives](./02_architecture/uml_diagrams/02_component_deep_dives.md)**: 5-Step Risk Model and Predictive Engine logic.
- **[Level 3: Traceability Map](./02_architecture/uml_diagrams/03_implementation_traceability_map.md)**: Visual bridge between domains and source code.
- **[Level 4: Function Micro-Flows](./02_architecture/uml_diagrams/04_function_level_deep_dives.md)**: Granular internal logic of key functions.
- **[Level 5: Implementation Blueprint](./02_architecture/uml_diagrams/05_implementation_blueprint.md)**: Annotated source code walk-through.

### Dual-Process Pipeline
1.  **MigraGuard Agent**: A lightweight background collector that harvests metrics from `pg_stat_statements` and stores time-series workload patterns in a local SQLite database.
2.  **MigraGuard Analyze**: A CLI tool that parses migration SQL into an AST (Abstract Syntax Tree) and evaluates risks using the stored patterns.

### Advanced Design Patterns
- **Strategy Pattern**: Isolates the 5-step risk assessment stages into atomized files under the dedicated `evaluators/` subpackage, successfully satisfying SRP and OCP.
- **Factory Pattern**: Delivers distinct **Live Mode** and **Sandbox Mode** environments for high reproducibility.
- **Dependency Injection**: Couples structured logger (`types.Logger`) and virtual database adapters seamlessly at the Client layer, securing complete domain decoupling.

---

## 📊 Mathematical Risk Model

The core of MigraGuard is a **5-Step Quantitative Model** derived from queuing theory and database internals:

1.  **Execution Time ($T_{ddl}$)**: Predicts I/O cost based on table size and storage throughput.
2.  **Blocking Window ($T_{block}$)**: $T_{ddl} + P99_{latency} + Lag_{repl}$.
3.  **Connection Peak ($C_{peak}$)**: Forecasts connection spikes during the blocking window: $C_{active} + (\lambda \times T_{block})$.
4.  **Recovery Cost ($T_{rec}$)**: Evaluates the system's ability to drain the backlog after the lock is released.
5.  **Risk Scoring**: Assigns **Safe / Warning / Danger** levels based on system capacity ($C_{max}, \mu_{max}$).

---

## 🏜️ Simulation Sandbox (Research Ready)

For academic validation and "what-if" analysis, MigraGuard provides a high-fidelity **Simulation Engine**. It generates months of realistic workload data (sine waves, event spikes, Gaussian noise) from a simple YAML scenario.

```bash
# Generate 7 days of simulated traffic
./build/migraguard simulate --scenario experiments/scenarios/05_spike_flash_sale.yaml

# Run analysis against the simulated world
./build/migraguard analyze migration.sql --sandbox exp_05_spike_flash_sale.db --forecast
```

The sandbox database is created in the current working directory as `<experiment_name>.db`, based on the `experiment_name` field in the scenario YAML.

---

## 🛠️ Quick Start

### Installation
```bash
go install github.com/Homeria/MigraGuard/cmd/migraguard@latest
```

### 1. Start Monitoring (Agent)
```bash
migraguard agent --db "postgres://user:pass@localhost:5432/db"
```

### 2. Analyze Migration (CLI)
```bash
migraguard analyze ./migrations/001_heavy_alter.sql
```

For offline experiments, pass a generated SQLite sandbox explicitly:

```bash
migraguard analyze ./migrations/001_heavy_alter.sql --sandbox exp_05_spike_flash_sale.db
```

---

## 📚 Documentation Hub

Detailed documentation is available in both Korean and English.

| Category | KR (한국어) | EN (English) |
| :--- | :--- | :--- |
| **Requirements** | [KR-01](./docs/kr/01_requirements/functional_spec.md) | [EN-01](./docs/en/01_requirements/functional_spec.md) |
| **Architecture** | [KR-02](./docs/kr/02_architecture/system_overview.md) | [EN-02](./docs/en/02_architecture/system_overview.md) |
| **Implementation** | [KR-03](./docs/kr/03_implementation/parser_logic.md) | [EN-03](./docs/en/03_implementation/parser_logic.md) |
| **User Guide** | [KR-04](./docs/kr/04_guides/sandbox_manual.md) | [EN-04](./docs/en/04_guides/sandbox_manual.md) |

---

## 📜 License
Distributed under the **MIT License**. Created by **Homeria / MigraGuard Team**.
