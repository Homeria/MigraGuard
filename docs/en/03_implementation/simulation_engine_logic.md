# 🧠 Simulation Sandbox Engine: Mathematical Principles

This document explains the statistical and mathematical logic behind the high-fidelity time-series data generation in MigraGuard v3.8.

---

## 1. Overview

The Simulation Sandbox Engine employs **Stochastic Workload Modeling** to replicate the complex patterns of real-world database traffic. The goal is to provide a dataset suitable for risk analysis validation and AI/ML model training.

---

## 2. Core Modeling Principles

### 2.1. Multi-Layer Periodicity (Daily & Weekly Cycles)
The engine uses a superposition of sine waves to model human activity patterns:
- **Primary Wave**: A 24-hour cycle mimicking daylight activity.
- **Secondary Wave**: Mid-day (14:00) and evening (20:00) peaks.
- **Weekly Adjustment**: Reduces traffic by 40% on weekends to simulate B2C patterns.

**Formula**:
$$TPS_{base}(t) = (Base + \text{SineDaily}(t) \times (Peak - Base)) \times \text{WeeklyMult}(t)$$

### 2.2. Metric Correlation
- **Active Connections ($C$):** Grows linearly with TPS.
- **P99 Latency ($L$):** Grows **exponentially** as TPS approaches system capacity ($\mu_{max}$), replicating queuing delays.

**Formula**:
$$L_{p99}(t) = L_{base} \times e^{(\frac{TPS(t)}{\mu_{peak}} - 1)}$$

### 2.3. Event-Driven Spikes (Anomaly Injection)
Timeline events (e.g., Flash Sales) apply a **Multiplier** to the base TPS calculation, testing engine sensitivity to sudden surges.

### 2.4. Stochastic Noise (Gaussian Variance)
Programmable noise variance ($\sigma$) is added at every 1-minute interval to ensure data realism.

---

## 3. Data Distribution across Tables

Load is distributed across the **Fintech Commerce Schema**:
- **`order_event_logs`**: 150% of base load (Write-heavy).
- **`orders` / `inventory_stocks`**: 100% of base load (Transaction-heavy).
- **`account_balances`**: 30% of base load (Read/Update-focused).
