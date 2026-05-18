# Level 4: Function-Level Micro-Flows (Rich Implementation Nodes)

This diagram details MigraGuard's core logic at the function level, describing exactly which file and function perform each step with what arguments.

---

## 1. `analyze` Command Execution Logic (Analyze Pipeline)

```mermaid
flowchart TD
    Step1["<b>File:</b> analyze_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> ctx, task<br/><b>Role:</b> Orchestrate the entire DDL analysis process"]
    
    Step2["<b>File:</b> sql_parser.go<br/><b>Func:</b> ParseSQL()<br/><b>Args:</b> sqlContent<br/><b>Role:</b> Convert SQL to AST and identify DDL types (Table/Index/Column)"]
    
    Step3["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeRisk()<br/><b>Args:</b> ctx, analysisResult<br/><b>Role:</b> Calculate 5-step risk score based on current DB state"]
    
    Step4["<b>File:</b> sqlite_analyzer.go<br/><b>Func:</b> Get24HourTrafficForecast()<br/><b>Args:</b> ctx, tableName<br/><b>Role:</b> Generate 24h traffic profile based on past 7-day metrics"]
    
    Step5["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeForecast()<br/><b>Args:</b> ctx, analysis, forecastData<br/><b>Role:</b> Run 24 virtual simulations to find the Golden Window"]
    
    Step6["<b>File:</b> analyze_service.go<br/><b>Func:</b> exportForecastCSV()<br/><b>Args:</b> forecastReports<br/><b>Role:</b> Export CSV data for visualization tool (Python)"]

    Step1 --> Step2
    Step2 --> Step3
    Step3 --> Step4
    Step4 --> Step5
    Step5 --> Step6
    
    style Step1 fill:#f9f,stroke:#333,stroke-width:2px
    style Step3 fill:#fff3e0,stroke:#e65100,stroke-width:2px
    style Step5 fill:#fff3e0,stroke:#e65100,stroke-width:2px
```

---

## 2. Predictive Engine Internal Logic (Forecast & Tie-Breaker)

```mermaid
flowchart TD
    F1["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeForecast()<br/><b>Args:</b> ctx, analysis, forecast<br/><b>Role:</b> Start 24-hour virtual simulation loop"]
    
    F2["<b>File:</b> risk_evaluator.go<br/><b>Func:</b> Evaluate()<br/><b>Args:</b> ctx, metrics, report, constants<br/><b>Role:</b> Execute 5-step model with virtual metrics (ExpectedTPS/P99)"]
    
    F3["<b>File:</b> risk_calculator.go<br/><b>Internal Logic</b><br/><b>Logic:</b> math.Abs(score - minScore) < 0.001<br/><b>Role:</b> Apply TPS Tie-breaker for identical risk scores"]
    
    F4["<b>File:</b> risk_calculator.go<br/><b>Func:</b> report.BestHour update<br/><b>Args:</b> currentHour<br/><b>Role:</b> Finalize Golden Window with lowest Risk & TPS"]

    F1 --> F2
    F2 --> F3
    F3 --> F4
    
    style F2 fill:#e1f5fe,stroke:#01579b
    style F3 fill:#ffecb3,stroke:#ff6f00,stroke-width:3px
```

---

## 3. Sandbox Data Generation Logic (Sandbox Seeding)

```mermaid
flowchart TD
    S1["<b>File:</b> simulate_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> ctx, scenarioPath, force<br/><b>Role:</b> Load scenario YAML and initialize Sandbox DB"]
    
    S2["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> SeedScenario()<br/><b>Args:</b> scenario<br/><b>Role:</b> Loop to generate 7-day (10,080 points) time-series data"]
    
    S3["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> CalculateTPS()<br/><b>Args:</b> time, profile<br/><b>Role:</b> Calculate TPS based on Sine Wave + Weekly + Noise"]
    
    S4["<b>File:</b> sqlite_sandbox.go<br/><b>Database Action</b><br/><b>Args:</b> metricStmt.Exec(...)<br/><b>Role:</b> Batch insert generated virtual metrics into SQLite"]

    S1 --> S2
    S2 --> S3
    S3 --> S4
    
    style S3 fill:#f1f8e9,stroke:#33691e
```
