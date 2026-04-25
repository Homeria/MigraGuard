# 🏛️ MigraGuard System Architecture Overview

This document describes the **SDK-First Architecture** of MigraGuard, its data flow, and the design philosophy of the core risk model.

---

## 1. SDK-First Modular Architecture

MigraGuard is designed as a reusable library (SDK) that can be embedded into various interfaces such as CLI tools, MCP servers, or CI/CD pipelines.

### 1.1. Core Package Structure
- **`pkg/migraguard` (The Entrance)**: The public entry point for all features. Provides the `Client` which orchestrates analysis and agent services.
- **`pkg/migraguard/types` (The Models)**: Shared data structures for analysis results and risk reports.
- **`pkg/migraguard/internal` (The Private Core)**: Encapsulated core domains:
    - **Analyzer**: `sql_parser.go`, `risk_calculator.go`, `risk_evaluator.go`.
    - **Collector**: `metric_collector.go` for background data harvesting.
    - **Infra**: Database adapters for PostgreSQL and SQLite.

### 1.2. Architecture Layers
1. **Interface Layer**: CLI (`cmd/`), MCP Server, or API.
2. **SDK Layer**: The public `Client` interface (`pkg/migraguard`).
3. **Domain Layer**: Core logic and algorithms (`internal/analyzer`, `collector`).
4. **Adapter Layer**: Communication with external systems (`internal/infra`).

---

## 2. Flexible Configuration System

MigraGuard uses a hierarchical configuration system (YAML > Env Vars > Defaults).

| Component | Setting File | Default Location |
| :--- | :--- | :--- |
| **Main Config** | `migraguard.yaml` | Current directory or `/etc/migraguard/` |
| **Overrides** | `.env` | Environment variables prefixed with `MIGRAGUARD_` |

---

## 3. 5-Step 정밀 리스크 모델 (5-Step Risk Model)

MigraGuard calculates the final risk score through the following 5 stages:

1.  **Step 1: $T_{ddl}$ (Schema Analysis)**: Predict execution time based on DDL type and rewrite requirement.
2.  **Step 2: $T_{block}$ (Lock Contention)**: Calculate potential service blocking time.
3.  **Step 3: $C_{peak}$ (Peak Concurrency)**: Predict maximum influx of connections.
4.  **Step 4: $T_{rec}$ (Recovery Time)**: Evaluate cost for system normalization.
5.  **Step 5: Risk Classification**: Determine Safe / Warning / Danger level based on configurable thresholds.

---

## 4. Maintenance and Implementation Details

- **Clean Root Strategy**: Infrastructure files (Dockerfile, init scripts) are kept in `build/`, and utility scripts in `scripts/`.
- **Role-based Naming**: All files follow the `{domain}_{role}.go` naming convention for maximum maintainability.
- **English Standard**: All comments and logs are standardized to English (ASCII) for cross-environment build stability.
