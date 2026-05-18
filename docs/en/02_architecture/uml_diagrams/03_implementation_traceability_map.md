# Level 3: Visual Implementation Traceability Map

This diagram provides a direct visual link between the architectural domains and the specific Go files and functions that implement them.

```mermaid
graph LR
    subgraph Interface_Layer [Interface Layer - CLI]
        direction TB
        CMD_A[analyze.go] -->|calls| FN_RA[runAnalyze]
        CMD_S[simulate.go] -->|calls| FN_RS[runSimulate]
        CMD_C[check.go] -->|calls| FN_RC[runCheck]
    end

    subgraph Service_Layer [Application Domain - Service]
        direction TB
        SVC_A[analyze_service.go] -->|Run| FN_AS_R[AnalyzeService.Run]
        SVC_S[simulate_service.go] -->|Run| FN_SS_R[SimulateService.Run]
        SVC_G[agent_service.go] -->|Run| FN_GS_R[AgentService.Run]
    end

    subgraph Core_Logic [Domain Logic - Risk Engine]
        direction TB
        CALC[risk_calculator.go] -->|Real-time| FN_AR[AnalyzeRisk]
        CALC -->|Prediction| FN_AF[AnalyzeForecast]
        EVAL[risk_evaluator.go] -->|Strategy| IF_SE[StepEvaluator Interface]
        PARS[sql_parser.go] -->|AST| FN_PS[ParseSQL]
    end

    subgraph Infra_Layer [Infrastructure Domain - Adapter]
        direction TB
        PG_AD[postgres/adapter.go] -->|Real-time| FN_FTD[FetchTableDynamicMetrics]
        SQL_RP[sqlite_repository.go] -->|Storage| FN_IS[InitializeSchema]
        SQL_VT[sqlite_virtual.go] -->|Virtualization| FN_VPG[VirtualPGAdapter]
        SQL_SB[sqlite_sandbox.go] -->|Seeding| FN_SS[SeedScenario]
    end

    %% Cross-Layer Call Streams
    FN_RA --> FN_AS_R
    FN_RS --> FN_SS_R
    FN_RC --> FN_GS_R

    FN_AS_R --> FN_AR
    FN_AS_R --> FN_AF
    FN_AS_R --> FN_PS
    
    FN_AR --> IF_SE
    FN_AF --> IF_SE

    FN_AR --> FN_FTD
    FN_SS_R --> FN_SS
    FN_AF --> SQL_VT

    style Interface_Layer fill:#f5f5f5,stroke:#333,stroke-dasharray: 5 5
    style Service_Layer fill:#e1f5fe,stroke:#01579b
    style Core_Logic fill:#fff3e0,stroke:#e65100,stroke-width:3px
    style Infra_Layer fill:#f1f8e9,stroke:#33691e
```

### Traceability Highlights
*   **Analysis Stream**: `analyze.go` (CLI) $\rightarrow$ `analyze_service.go` (Orchestration) $\rightarrow$ `risk_calculator.go` (Core Math).
*   **Mathematical Execution**: The `AnalyzeRisk` and `AnalyzeForecast` functions in `risk_calculator.go` act as the brain, calling individual `StepEvaluator` strategies in `risk_evaluator.go`.
*   **Simulation Stream**: `simulate.go` (CLI) $\rightarrow$ `simulate_service.go` (Orchestration) $\rightarrow$ `sqlite_sandbox.go` (Data Generation).
*   **Virtualization**: When in sandbox mode, the `AnalyzeForecast` logic switches from the real `postgres/adapter.go` to the `sqlite_virtual.go` adapter to avoid touching production databases.
