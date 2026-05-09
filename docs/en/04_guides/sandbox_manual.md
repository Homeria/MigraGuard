# 🏜️ MigraGuard Simulation Sandbox Manual

This manual describes the usage and mathematical principles of the **Simulation Sandbox**, which builds virtual workload environments for risk validation without a physical PostgreSQL connection.

---

## 1. Overview

The simulation sandbox is designed for **Research Reproducibility**. It generates 7 days of time-series workload data in seconds using mathematical models, providing a robust offline experimentation environment.

### Core Values
*   **Offline Validation**: Test all risk engine scenarios without any DB connection.
*   **Reproducibility**: Ensure identical experiment results globally using the same YAML scenario.
*   **Edge Case Simulation**: Freely design scenarios like "Flash Sales" or "System Degradation" that are hard to reproduce in production.

---

## 2. Scenario Definition (YAML Schema)

Key fields for the experiment blueprint.

| Field | Description | Note |
| :--- | :--- | :--- |
| `target_table` | Name of the primary table for analysis. | Required |
| `pg_state` | Mock production state at the time of analysis (Size, P99, etc.). | Required |
| `base_tps` / `peak_tps` | Baseline and maximum traffic loads. | Required |
| `weekly_pattern` | Whether to apply a 40% traffic reduction on weekends. | Boolean |
| `events` | Definition of sudden traffic spikes (e.g., Flash Sales). | Array |

---

## 3. Mathematical Workload Models

The sandbox engine uses the following formulas to generate realistic operational data.

### 3.1. Daily/Weekly Traffic Model
The TPS ($\lambda$) at time $t$ is determined by:
$$\lambda(t) = (\lambda_{base} + \text{Sine}(t) \times (\lambda_{peak} - \lambda_{base})) \times M_{week} \times M_{event} + \epsilon$$
*   $\text{Sine}(t)$: A dual-sine wave model peaking at 14:00 and 20:00.
*   $\epsilon$: Gaussian-like noise for realistic irregularity.

### 3.2. Metric Correlation Model
*   **Latency (P99)**: Increases exponentially as TPS rises.
    $$P99 = P99_{base} \times e^{(\frac{\lambda}{\lambda_{peak}} - 1)}$$
*   **Active Connections**: Increases linearly in proportion to TPS.

---

## 4. CLI Usage

### 4.1. Generating a Sandbox
```bash
migraguard simulate --scenario scenario.yaml
```

### 4.2. Offline Risk Analysis
```bash
migraguard analyze migration.sql --sandbox scenario.db
```

---

## 5. Use Cases
- Pre-deployment risk gatekeeping in CI/CD pipelines.
- Academic research on risk score trends relative to TPS and data volume.
- Lightweight education and demo environments without heavy DB infra.
