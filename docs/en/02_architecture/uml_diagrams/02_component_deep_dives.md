# Level 2: Component Deep-Dives

This document explores the internal logic of the Risk Engine and the Golden Window discovery algorithm.

## 1. The 5-Step Quantitative Risk Model

Each DDL is evaluated through five systematic stages to calculate the final risk score.

```mermaid
sequenceDiagram
    participant P as PostgreSQL/Sandbox
    participant E as Risk Engine
    participant R as Analysis Report

    E->>P: Fetch Metrics (TPS, P99, TableSize, ActiveConns)
    Note over E: Step 1: T_ddl (Predict duration)
    E->>E: Predict duration based on DiskIO vs Size
    
    Note over E: Step 2: T_block (Lock duration)
    E->>E: T_block = (P99 + T_ddl + Lag) * LockImpact
    
    Note over E: Step 3: C_peak (Concurrency surge)
    E->>E: C_peak = ActiveConns + (TPS/1000 * T_block)
    
    Note over E: Step 4: T_rec (Recovery time)
    E->>E: T_rec = (C_peak - C_max) / (Mu_max - TPS)
    
    Note over E: Step 5: Risk score & level
    E->>E: Score = (C_peak / C_max) * 100
    E->>R: Populate Report (Safe/Warning/Danger)
```

## 2. Golden Window Discovery (Predictive Forecast)

When `--forecast` is enabled, the engine performs 24 virtual simulations to find the safest time for deployment.

```mermaid
flowchart TD
    Start[Fetch 24h Forecast Profile] --> Loop{{For each hour 00..23}}
    Loop --> Virtual[Create Virtual Metrics Snapshot]
    Virtual --> Run[Run 5-Step Model for this slot]
    Run --> Score[Calculate RiskScore & ExpectedTPS]
    
    Score --> Better{Is this hour better?}
    
    Better -- "Score < MinScore" --> Update[Update BestHour]
    Better -- "Score == MinScore AND TPS < MinTPS" --> Update
    Better -- Else --> Next[Continue Loop]
    
    Update --> Next
    Next --> Loop
    Loop -- End --> Result([Identify Final Golden Window])
```

### The TPS Tie-Breaker Logic (v3.8 Improvement)
If multiple hours are deemed equally "Safe" by the risk engine (e.g., they all hit the *Base Risk* floor of 30.0), the engine selects the hour with the **lowest Expected TPS**. This ensures the deployment is scheduled during the period of absolute minimum physical traffic, maximizing the safety margin.
