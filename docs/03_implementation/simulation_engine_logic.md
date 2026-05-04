# 🧠 Simulation Sandbox Engine: Mathematical Principles

This document explains the statistical and mathematical logic behind the high-fidelity time-series data generation in MigraGuard v3.8.

---

## 1. Overview

The Simulation Sandbox Engine does not generate random numbers. Instead, it employs **Stochastic Workload Modeling** to replicate the complex patterns of real-world enterprise database traffic. The goal is to provide a dataset suitable for both risk analysis validation and AI/ML model training.

---

## 2. Core Modeling Principles

### 2.1. Multi-Layer Periodicity (Daily & Weekly Cycles)
Real traffic follows human behavior. The engine uses a superposition of sine waves to model this:
- **Primary Wave**: A 24-hour cycle mimicking daylight activity.
- **Secondary Wave**: A mid-day peak (lunchtime) and an evening peak (prime time).
- **Weekly Adjustment**: A coefficient that reduces traffic by 40% on weekends to simulate business-to-consumer (B2C) patterns.

**Formula**:
$$TPS_{base}(t) = (Base + SineDaily(t) \times (Peak - Base)) \times WeeklyMult(t)$$

### 2.2. Metric Correlation (The "Cascading Load" Effect)
In a real DB, metrics are interdependent. The engine simulates this through functional mapping:
- **Active Connections ($C$):** Grows linearly with TPS. As more requests arrive, the number of held connections increases.
- **P99 Latency ($L$):** Grows **exponentially** as TPS approaches the system's capacity ($\mu_{max}$). This replicates the "Queuing Delay" effect where response times degrade non-linearly under heavy load.

**Formula**:
$$L_{p99}(t) = L_{base} \times e^{(TPS(t)/\mu_{peak} - 1)}$$

### 2.3. Event-Driven Spikes (Anomaly Injection)
Researchers can inject "Timeline Events" (e.g., Flash Sales). During these windows, the engine applies a **Multiplier** to the base TPS calculation, creating outliers that test the engine's sensitivity to sudden traffic surges.

### 2.4. Stochastic Noise (Gaussian Variance)
To ensure the data isn't "too perfect" for AI training, a programmable noise variance ($\sigma$) is added at every data point (1-minute interval).
$$TPS_{final} = TPS_{calculated} \times (1 + \text{Uniform}(-1, 1) \times \text{NoiseVariance})$$

---

## 3. Data Distribution across Tables

The engine simulates a holistic system by distributing the total load across the **Fintech Commerce Schema**:
- **`order_event_logs`**: Receives 150% of the base load (Write-heavy).
- **`orders` / `inventory_stocks`**: Receives 100% of the base load (Transaction-heavy).
- **`account_balances`**: Receives 30% of the base load (Read/Update-focused).

---

## 4. Research Applications

1.  **Golden Window Identification**: By generating 7 days of cyclical data, researchers can verify if `IdentifySafestDeploymentWindow()` correctly finds the 03:00 AM low-traffic slot.
2.  **Capacity Planning**: By setting $PeakTPS$ close to $\mu_{max}$, researchers can observe the threshold where `RecoveryTime` ($T_{rec}$) becomes infinite.
3.  **AI Training**: The 10,080 data points (1-min intervals for 7 days) provide a high-quality labeled dataset for training anomaly detection or forecasting models.
