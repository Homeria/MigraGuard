# 🔮 Predictive Analysis Engine: Forecasting Logic

This document details the implementation of the 24-hour predictive risk forecasting engine in MigraGuard v3.8.

---

## 1. Concept: From Retroactive to Predictive

While standard analysis focuses on the "current" state of the database, the **Predictive Engine** answers the question: *"When is the safest time to execute this DDL in the next 24 hours?"*

It achieves this by synthesizing historical workload patterns into a future forecast and running multiple "What-If" simulations.

---

## 2. Technical Workflow

### 2.1. Historical Profiling (Aggregation)
The engine queries the SQLite metric store to build a baseline profile. It aggregates data from the past 7 days, grouping metrics by the hour of the day.

**SQL Logic**:
```sql
SELECT 
    CAST(substr(timestamp, 12, 2) AS INTEGER) as hour, 
    AVG(tps) as avg_tps, 
    MIN(tps) as min_tps,
    MAX(tps) as max_tps,
    AVG(p99_time) as avg_p99 
FROM table_metrics 
WHERE table_name = ? 
GROUP BY hour;
```

### 2.2. In-Memory "What-If" Simulation Loop
Instead of a single risk calculation, the `RiskEngine` iterates through the 24 forecasted time slots.
- **Input**: For each hour $H \in \{0..23\}$, the engine injects $TPS_{expected}(H)$ and $Latency_{expected}(H)$ into the mathematical risk model.
- **Optimization**: Static metrics like `TableSize` are fetched once, while dynamic metrics are swapped in-memory to ensure near-instant calculation speeds (<50ms for 24 cycles).

---

## 3. Visualization: Risk Heatmap Overlay

The results are exported to `predictive_forecast.csv` and processed by a Python-based visualizer (`plot_predictive_heatmap.py`).

### 3.1. Variance Cloud (Min-Max Shading)
To convey prediction uncertainty, the graph displays a "Variance Cloud" around the average TPS line using the `MinTPS` and `MaxTPS` values.
- **Line**: Represents the expected (average) workload.
- **Shaded Area**: Represents the historical range of traffic fluctuations.

### 3.2. Risk-Based Background Coloring
The plot area is color-coded using `axvspan` to provide immediate visual feedback:
- 🟢 **Safe Zone**: Predicted Risk Score < 50%
- 🟡 **Warning Zone**: 50% ≤ Predicted Risk Score < 80%
- 🔴 **Danger Zone**: Predicted Risk Score ≥ 80% (or connection pool overflow)

---

## 4. Automated Decision Support

The engine automatically identifies the **Golden Window**—the hour with the absolute lowest risk score across the 24-hour period—and highlights it with a star (★) annotation on the final report.
