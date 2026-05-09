# 🏛️ MigraGuard System Architecture Overview

This document describes the **SDK-First Architecture** of MigraGuard, its data flow, and the design philosophy of the core risk model.

---

## 1. SDK-First Modular Architecture

MigraGuard is designed as a reusable library (SDK) that can be embedded into various interfaces such as CLI tools, MCP servers, or CI/CD pipelines.

### 1.1. Core Package Structure
- **`pkg/migraguard` (Entry)**: The public entry point. Provides the `Client` with Factory Methods (`NewLiveClient`, `NewSandboxClient`) for explicit environment separation.
- **`pkg/migraguard/types` (Models)**: Shared data structures, risk reports, and interface definitions.
- **`pkg/migraguard/internal` (Private Core)**: Encapsulated domain logic:
    - **Analyzer**: Modular risk assessment using the **Strategy Pattern** (`StepEvaluator`) for modular $T_{ddl}, T_{block}$ calculations.
    - **Collector**: Background metric harvesting via `metric_collector.go`.
    - **Infra**: Database adapters for PostgreSQL and SQLite with consistent **Context Propagation**.
    - **Shared**: Standardized **Error Coding System** (`MG-XXX-###`).

### 1.2. Architecture Layers
1. **Interface Layer**: CLI (`cmd/`), MCP Server, or API.
2. **SDK Layer**: High-level `Client` API with Dependency Injection (DI) support.
3. **Domain Layer**: Orchestrated via `internal/app` (Service Layer Pattern).
4. **Adapter Layer**: Logic-less database drivers and simulation virtualization (`VirtualPGAdapter`).

---

## 2. Flexible Configuration System

MigraGuard uses a hierarchical configuration system (YAML > Env Vars > Defaults).

| Component | Setting File | Default Location |
| :--- | :--- | :--- |
| **Main Config** | `migraguard.yaml` | Current directory or `/etc/migraguard/` |
| **Overrides** | `.env` | Environment variables prefixed with `MIGRAGUARD_` |

---

## 3. 5-Step Quantitative Risk Model

MigraGuard calculates the final risk score through 5 systematic stages based on operational data:

1.  **Step 1: $T_{ddl}$ (Schema Analysis)**: Predicts execution time based on DDL type and table size.
2.  **Step 2: $T_{block}$ (Lock Contention)**: Calculates the potential blocking window by summing DDL time, P99 latency, and replication lag.
3.  **Step 3: $C_{peak}$ (Peak Concurrency)**: Forecasts the maximum connection surge during the blocking window.
4.  **Step 4: $T_{rec}$ (Recovery Cost)**: Evaluates the time required for system normalization after the lock release.
5.  **Step 5: Risk Classification**: Assigns Safe / Warning / Danger levels based on system capacity thresholds ($C_{max}, \mu_{max}$).

---

## 4. Research Workspace

The `experiments/` directory provides a structured environment for offline validation:
- `scenarios/`: YAML-based declarative load profiles.
- `ddl/`: Research-specific migration SQL cases.
- `data/`: Persistence layer for generated SQLite sandboxes.
