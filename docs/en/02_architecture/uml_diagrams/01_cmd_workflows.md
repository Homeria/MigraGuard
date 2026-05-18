# Level 1: Command-Level Workflows

This document details the operational flow for each primary MigraGuard command.

## 1. `analyze` Workflow (Risk Assessment & Forecast)

The `analyze` command is the core "Circuit Breaker" of the system.

```mermaid
flowchart TD
    Start([Start]) --> Parse[Parse SQL File - AST Analysis]
    Parse --> Mode{Analysis Mode?}
    
    Mode -- Live --> PG[Fetch Metrics from Live PostgreSQL]
    Mode -- Sandbox --> SL[Fetch Metrics from SQLite Sandbox]
    
    PG & SL --> Engine[Execute 5-Step Risk Engine]
    
    Engine --> Forecast{--forecast enabled?}
    
    Forecast -- No --> Report[Generate Console/CSV Report]
    
    Forecast -- Yes --> History[Fetch 24h Traffic Baseline from SQLite]
    History --> Loop[Iterate 24 Hours - Simulate Risk]
    Loop --> TieBreak[Identify Golden Window - Lowest Risk & TPS]
    TieBreak --> Visual[Python: Generate Heatmap PNG]
    Visual --> Report
    
    Report --> Decision{Risk > Threshold?}
    Decision -- Danger --> Block([Block Pipeline - Exit 1])
    Decision -- Safe --> Allow([Allow Pipeline - Exit 0])
```

## 2. `simulate` Workflow (Research Sandbox Generation)

Allows researchers to create controlled environments via declarative YAML.

```mermaid
flowchart TD
    Start([Start]) --> YAML[Read Scenario YAML]
    YAML --> Prof[Initialize Workload Profiler]
    Prof --> DB[Create fresh {Experiment}.db]
    DB --> Seed[Seed 7-day History with Sine/Noise Patterns]
    Seed --> Finish([Sandbox Ready])
```

## 3. `agent` Workflow (Background Collection)

Ensures the system has enough historical data for predictive forecasting.

```mermaid
flowchart TD
    Start([Start]) --> Loop[Interval Tick - e.g., every 1m]
    Loop --> Harvest[Fetch pg_stat_statements & Table Sizes]
    Harvest --> Save[UPSERT into SQLite metrics table]
    Save --> Purge[Cleanup data older than retention days]
    Purge --> Loop
```
