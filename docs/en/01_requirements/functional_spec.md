# 🛡️ MigraGuard Functional Specification

## 1. Project Definition
**MigraGuard** is a **Traffic-Aware Database Guardrail** that detects potential service disruption risks during PostgreSQL schema changes (DDL). It cross-analyzes real-time traffic patterns with the static characteristics of SQL to verify deployment safety.

## 2. Core Values and Goals
- **Precision Load Tracking**: Calculates accurate time-point TPS by measuring deltas in cumulative metrics.
- **High-Fidelity Simulation**: Provides an environment that mimics 24-hour traffic cycles and business scenarios.
- **Conservative Risk Model**: A 5-step evaluation engine that dynamically adjusts based on real-time load intensity.
- **Automated Gatekeeping**: Prevents outages by blocking CI/CD pipelines when high risk (Danger) is detected.
- **Rapid Logic Validation**: An injection environment enabling instant simulation without waiting days for data collection.

## 3. Key Feature Details (v3.8)

### 3.1. Scenario-Based Data Seeder
- **Instant Data Generation**: Injects 7 to 30 days of virtual time-series data into SQLite in seconds.
- **Diverse Load Scenarios**:
    - `Normal`: Regular traffic patterns with consistent cycles.
    - `Flash Sale`: Sudden spikes in TPS (over 10x) at specific intervals.
    - `Degradation`: Gradual decline in disk I/O performance simulating system aging.

### 3.2. High-Fidelity Load Generator
- **Domain-Driven Load**: Reproduces complex transactional states in E-commerce domains (Orders, Inventory, Products).
- **Traffic Curve Application**: Uses sine wave formulas to realistically simulate hourly traffic fluctuations.

### 3.3. Intelligent Collection Agent (Background Agent)
- **Metric Harvesting**: Collects per-table performance metrics using `pg_stat_statements`.
- **Automated Storage Management**: Periodic garbage collection (Retention) to optimize SQLite storage capacity.

### 3.4. Dynamic Risk Engine
- **Strategy Pattern Integration**: Decouples analysis stages ($T_{ddl}, T_{block}$, etc.) into independent objects for extensibility.
- **5-Step Quantitative Report**: Provides metrics ($T_{ddl}, T_{block}, C_{peak}, T_{rec}$) based on queuing theory and database internals.
- **Golden Window Recommendation**: Suggests the safest deployment window by analyzing the last 24 hours of data.
