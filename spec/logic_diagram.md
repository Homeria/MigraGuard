# 📊 MigraGuard v3.2 아키텍처 및 로직 다이어그램

본 문서는 MigraGuard의 서비스 흐름을 시각화하기 위한 Mermaid 다이어그램 코드를 담고 있습니다.

## 1. 전체 레이어 구조 (Architecture Layers)

```mermaid
graph TD
    subgraph CLI_Layer [CLI & Presentation Layer]
        Main[main.go] --> Root[root.go]
        Root --> AgentCmd[agent.go]
        Root --> AnalyzeCmd[analyze.go]
        AnalyzeCmd --> Reporter[Reporter Interface]
    end

    subgraph Service_Layer [Domain Service Layer]
        AgentCmd --> AgentSvc[AgentService]
        AnalyzeCmd --> AnalyzeSvc[AnalyzeService]
        AnalyzeSvc --> RiskEngine[RiskEngine]
    end

    subgraph Infrastructure_Layer [Infrastructure & Adapters]
        AgentSvc --> Collector[Collector]
        AnalyzeSvc --> Parser[SQL Parser]
        Collector --> PG[Postgres Adapter]
        Collector --> SL[SQLite Adapter]
        RiskEngine --> PG
        RiskEngine --> SL
    end

    subgraph External [External Resources]
        PG --> Postgres[(PostgreSQL DB)]
        SL --> SQLite[(Local SQLite File)]
    end
```

---

## 2. Agent 모드 수집 흐름 (Agent Collection Flow)

```mermaid
sequenceDiagram
    participant OS as OS/User
    participant AG as AgentCmd
    participant SVC as AgentService
    participant COL as Collector
    participant PG as PostgresAdapter
    participant SL as SQLiteAdapter

    OS->>AG: migraguard agent --db...
    AG->>SVC: NewAgentService(pg, sqlite)
    AG->>SVC: Run(ctx)
    SVC->>COL: Start(ticker)
    
    loop Every Interval
        COL->>PG: FetchWorkload()
        PG-->>COL: []WorkloadSnapshot
        COL->>SL: SaveSnapshots(data)
        
        loop For Target Tables
            COL->>PG: GetTableDynamicMetrics(table)
            PG-->>COL: *TableDynamicMetrics
            COL->>SL: SaveTableMetrics(metrics)
        end
        
        COL->>SL: PurgeOldSnapshots(retention)
        SL->>SL: VACUUM
    end
```

---

## 3. Analyze 모드 분석 흐름 (Analyze Risk Flow)

```mermaid
sequenceDiagram
    participant OS as OS/User
    participant AL as AnalyzeCmd
    participant SVC as AnalyzeService
    participant PS as SQL Parser
    participant RE as RiskEngine
    participant PG as PostgresAdapter
    participant SL as SQLiteAdapter
    participant RP as Reporter

    OS->>AL: migraguard analyze migration.sql
    AL->>SVC: Run(AnalysisTask)
    
    SVC->>PS: ParseSQL(content)
    PS-->>SVC: []AnalysisResult (Table, Lock, F_rewrite)
    
    loop For Each Table
        SVC->>PG: ValidateSchema(table, cols)
        SVC->>RE: AnalyzeRisk(analysis)
        
        RE->>PG: GetTableDynamicMetrics(table)
        RE->>SL: GetRecentTPSDelta(table)
        RE->>SL: GetTableBaselineStats(table)
        
        Note over RE: Calculate Risk Score<br/>Max(Current, Avg, Peak)
        
        RE-->>SVC: *RiskAnalysisReport
    end
    
    SVC-->>AL: *AnalysisResponse
    AL->>RP: Write(results, reports)
    
    alt RiskLevel == "Danger"
        AL->>OS: Exit(1)
    else
        AL->>OS: Exit(0)
    end
```

---

## 5. 초정밀 함수 호출 그래프 (Function-Level Call Graph)

본 다이어그램은 MigraGuard v3.2의 모든 Go 파일 내 주요 함수들 간의 호출 관계(Caller -> Callee)를 상세히 보여줍니다.

```mermaid
flowchart TD
    %% [Layer 1] Main Entry & Root
    subgraph Main_Entry [main.go & cmd/root.go]
        M1[main] --> M2[cmd.Execute]
        M2 --> R1[rootCmd.Execute]
        R1 -.-> R2[initConfig: viper load config]
    end

    %% [Layer 2] Agent Mode Call Stack
    subgraph Agent_Mode [cmd/migraguard/agent.go]
        R1 --> AG1[agentCmd.Run]
        AG1 --> AG2[db.NewPostgresAdapter]
        AG1 --> AG3[db.NewSQLiteAdapter]
        AG1 --> AG4[service.NewAgentService]
        AG1 --> AG5[svc.Run: Context based loop]
    end

    subgraph Agent_Service [internal/service/agent.go]
        AG5 --> AS1[AgentService.Run]
        AS1 --> AS2[db.NewCollector]
        AS1 --> AS3[collector.Start]
        AS1 -- Wait --> AS4[<-ctx.Done: Graceful Shutdown]
    end

    subgraph Collector_Logic [internal/db/collector.go]
        AS3 --> DC1[collector.Start loop: ticker.C]
        DC1 --> DC2[collector.collect]
        
        DC2 --> PG1[pg.FetchWorkload]
        PG1 -- returns []WorkloadSnapshot --> DC2
        
        DC2 --> SL1[sqlite.SaveSnapshots]
        
        %% Loop representation for Target Tables
        DC2 --> LP1{For Each Table}
        LP1 --> PG2[pg.GetTableDynamicMetrics]
        PG2 -- returns *TableDynamicMetrics --> LP1
        LP1 --> SL2[sqlite.SaveTableMetrics]
        SL2 --> LP1
        
        LP1 --> SL3[sqlite.PurgeOldSnapshots]
        SL3 --> SL4[sqlite.VACUUM]
    end

    %% [Layer 3] Analyze Mode Call Stack
    subgraph Analyze_Mode [cmd/migraguard/analyze.go]
        R1 --> AZ1[analyzeCmd.Run]
        AZ1 --> AZ2[db.NewPostgresAdapter]
        AZ1 --> AZ3[db.NewSQLiteAdapter]
        AZ1 --> AZ4[service.NewAnalyzeService]
        AZ1 --> AZ5[svc.Run: AnalysisTask]
        
        AZ5 -- returns *AnalysisResponse --> AZ1
        
        AZ1 --> RP1[reporter.New...Reporter]
        AZ1 --> RP2[rpt.Write: results, reports]
        
        AZ1 --> AZ6{hasDanger?}
        AZ6 -- Yes --> AZ7[os.Exit 1]
        AZ6 -- No --> AZ8[os.Exit 0]
    end

    subgraph Analyze_Service [internal/service/analyze.go]
        AZ5 --> AVS1[AnalyzeService.Run]
        AVS1 --> AVS2[os.ReadFile: Load SQL]
        AVS1 --> PS1[parser.ParseSQL]
        
        %% Loop representation for results
        PS1 --> LP2{For Each Result}
        LP2 --> PG3[pg.ValidateSchema]
        LP2 --> RE1[riskEngine.AnalyzeRisk]
        RE1 --> LP2
        LP2 --> AZ5
    end

    subgraph SQL_Parser [internal/parser/ast.go]
        PS1 --> PS2[pg_query.Parse: Build AST]
        PS2 --> PS3[handleNode: switch stmt type]
        PS3 -- returns AnalysisResult --> PS1
    end

    subgraph Risk_Engine [internal/engine/risk.go]
        RE1 --> RE2[pg.GetTableDynamicMetrics]
        RE1 --> SL5[sqlite.GetRecentTPSDelta]
        RE1 --> SL6[sqlite.GetTableBaselineStats]
        RE1 --> SL7[sqlite.GetSafeWindow]
        
        RE1 --> RE3[Calculate Risk Metrics:<br/>Lambda_final, T_ddl, T_block, C_peak, RiskScore]
        
        RE3 -- returns *RiskAnalysisReport --> AVS1
    end

    %% [Layer 4] Database Adapters (Concrete Implementation)
    subgraph DB_Adapters [internal/db/...]
        PG1 & PG2 & PG3 --> Postgres_Client[(Postgres Instance)]
        SL1 & SL2 & SL4 & SL5 & SL6 & SL7 --> SQLite_File[(SQLite File)]
    end

    %% Styling
    style Agent_Mode fill:#e1f5fe,stroke:#01579b
    style Analyze_Mode fill:#fff3e0,stroke:#e65100
    style Collector_Logic fill:#f1f8e9,stroke:#33691e
    style Risk_Engine fill:#fce4ec,stroke:#880e4f
```
