# Level 1: CLI Command Implementation Flows

This document visualizes the execution paths for all CLI commands supported by MigraGuard. Each node includes the actual implementation file and function info.

---

## 1. `analyze` Command (Live / Sandbox / Forecast)
Analyzes the risk of a user's DDL and predicts future traffic.

```mermaid
flowchart TD
    A1["<b>File:</b> root.go / analyze.go<br/><b>Func:</b> runAnalyze()<br/><b>Args:</b> args, flags<br/><b>Role:</b> Validate CLI inputs and call AnalyzeService"]
    
    A2["<b>File:</b> analyze_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> task (SQLPath, Forecast)<br/><b>Role:</b> Orchestrate the entire analysis pipeline"]
    
    A3["<b>File:</b> sql_parser.go<br/><b>Func:</b> ParseSQL()<br/><b>Args:</b> sqlContent<br/><b>Role:</b> SQL AST analysis and DDL metadata extraction"]
    
    subgraph Execution_Path [Analysis Strategy Selection]
        A4_Live["<b>File:</b> postgres/adapter.go<br/><b>Func:</b> FetchTableDynamicMetrics()<br/><b>Args:</b> tableName<br/><b>Role:</b> Extract real-time metrics from live DB"]
        
        A4_Sandbox["<b>File:</b> sqlite_virtual.go<br/><b>Func:</b> FetchTableDynamicMetrics()<br/><b>Args:</b> tableName<br/><b>Role:</b> Extract virtual metrics from SQLite sandbox"]
    end
    
    A5["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeRisk()<br/><b>Args:</b> ctx, analysisResult<br/><b>Role:</b> Execute 5-step quantitative risk model"]
    
    A6["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeForecast()<br/><b>Args:</b> ctx, forecastData<br/><b>Role:</b> 24h virtual simulation loop (with Tie-breaker)"]
    
    A7["<b>File:</b> analyze_service.go<br/><b>Func:</b> exportForecastCSV()<br/><b>Args:</b> forecastReports<br/><b>Role:</b> Generate CSV for visualization"]

    A1 --> A2
    A2 --> A3
    A3 --> A4_Live
    A3 --> A4_Sandbox
    A4_Live & A4_Sandbox --> A5
    A5 -->|if --forecast| A6
    A6 --> A7
    
    style A5 fill:#fff3e0,stroke:#e65100,stroke-width:2px
    style A6 fill:#fff3e0,stroke:#e65100,stroke-width:2px
```

---

## 2. `simulate` Command (Sandbox Generation)
Generates virtual traffic scenarios to build a research environment.

```mermaid
flowchart TD
    S1["<b>File:</b> simulate.go<br/><b>Func:</b> runSimulate()<br/><b>Args:</b> scenarioPath, force<br/><b>Role:</b> Check scenario file path and call SimulateService"]
    
    S2["<b>File:</b> simulate_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> scenarioPath, force<br/><b>Role:</b> Parse YAML and manage sandbox DB file (.db)"]
    
    S3["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> SeedScenario()<br/><b>Args:</b> scenario<br/><b>Role:</b> Manage 7-day virtual data generation transaction"]
    
    S4["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> CalculateTPS()<br/><b>Args:</b> time, profile<br/><b>Role:</b> Calculate TPS based on math model (Sine/Weekly/Events)"]

    S1 --> S2
    S2 --> S3
    S3 --> S4
    
    style S3 fill:#f1f8e9,stroke:#33691e,stroke-width:2px
```

---

## 3. `check` Command (Background Agent)
Continuously collects live DB metrics and stores them in SQLite.

```mermaid
flowchart TD
    C1["<b>File:</b> check.go<br/><b>Func:</b> runCheck()<br/><b>Args:</b> targetTables<br/><b>Role:</b> Configure and start agent service"]
    
    C2["<b>File:</b> agent_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> ctx<br/><b>Role:</b> Infinite loop running at specified interval (1m)"]
    
    C3["<b>File:</b> postgres/adapter.go<br/><b>Func:</b> FetchCurrentWorkloadSnapshot()<br/><b>Args:</b> ctx<br/><b>Role:</b> Collect query stats from pg_stat_statements view"]
    
    C4["<b>File:</b> sqlite_repository.go<br/><b>Func:</b> UpdateWorkloadSnapshots()<br/><b>Args:</b> snapshots<br/><b>Role:</b> Store collected data in SQLite (UPSERT)"]
    
    C5["<b>File:</b> sqlite_init.go<br/><b>Func:</b> MaintenancePurgeData()<br/><b>Args:</b> retentionDays<br/><b>Role:</b> Cleanup old data exceeding retention period"]

    C1 --> C2
    C2 --> C3
    C3 --> C4
    C4 --> C5
    C5 -->|Loop| C2
    
    style C2 fill:#e1f5fe,stroke:#01579b,stroke-width:2px
```

---

## 4. `export` Command (Data Extraction)
Exports collected or generated sandbox data to research-grade CSV.

```mermaid
flowchart TD
    E1["<b>File:</b> export.go<br/><b>Func:</b> runExport()<br/><b>Args:</b> outputPath<br/><b>Role:</b> Verify output path and call client method"]
    
    E2["<b>File:</b> client.go<br/><b>Func:</b> ExportSandboxMetrics()<br/><b>Args:</b> outputPath<br/><b>Role:</b> Create file stream and orchestrate export"]
    
    E3["<b>File:</b> sqlite_repository.go<br/><b>Func:</b> FetchAllTableMetrics()<br/><b>Args:</b> ctx<br/><b>Role:</b> Query all stored metric data from SQLite"]
    
    E4["<b>File:</b> client.go<br/><b>Func:</b> Write CSV Rows<br/><b>Args:</b> metrics, writer<br/><b>Role:</b> Save queried data to file in CSV format"]

    E1 --> E2
    E2 --> E3
    E3 --> E4
    
    style E2 fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
```
