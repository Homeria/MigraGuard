# Level 3: 시각적 구현 추적성 맵 (Traceability Map)

이 다이어그램은 아키텍처 도메인과 이를 구현하는 구체적인 Go 파일 및 함수 간의 직접적인 연결 고리를 제공합니다.

```mermaid
graph LR
    subgraph Interface_Layer [인터페이스 레이어 - CLI]
        direction TB
        CMD_A[analyze.go] -->|호출| FN_RA[runAnalyze]
        CMD_S[simulate.go] -->|호출| FN_RS[runSimulate]
        CMD_C[check.go] -->|호출| FN_RC[runCheck]
    end

    subgraph Service_Layer [애플리케이션 레이어 - Service]
        direction TB
        SVC_A[analyze_service.go] -->|실행| FN_AS_R[AnalyzeService.Run]
        SVC_S[simulate_service.go] -->|실행| FN_SS_R[SimulateService.Run]
        SVC_G[agent_service.go] -->|실행| FN_GS_R[AgentService.Run]
    end

    subgraph Core_Logic [도메인 로직 - 리스크 엔진]
        direction TB
        CALC[risk_calculator.go] -->|실시간| FN_AR[AnalyzeRisk]
        CALC -->|예측| FN_AF[AnalyzeForecast]
        EVAL[risk_evaluator.go] -->|전략 패턴| IF_SE[StepEvaluator Interface]
        PARS[sql_parser.go] -->|AST 분석| FN_PS[ParseSQL]
    end

    subgraph Infra_Layer [인프라 레이어 - Adapter]
        direction TB
        PG_AD[postgres/adapter.go] -->|실시간 수집| FN_FTD[FetchTableDynamicMetrics]
        SQL_RP[sqlite_repository.go] -->|영속성 관리| FN_IS[InitializeSchema]
        SQL_VT[sqlite_virtual.go] -->|가상화| FN_VPG[VirtualPGAdapter]
        SQL_SB[sqlite_sandbox.go] -->|시나리오 시딩| FN_SS[SeedScenario]
    end

    %% 레이어 간 호출 흐름
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

### 추적성 하이라이트
*   **분석 흐름**: `analyze.go` (CLI) $\rightarrow$ `analyze_service.go` (오케스트레이션) $\rightarrow$ `risk_calculator.go` (핵심 수학 모델).
*   **수학적 연산**: `risk_calculator.go`의 `AnalyzeRisk`와 `AnalyzeForecast` 함수가 두뇌 역할을 하며, `risk_evaluator.go`에 정의된 개별 `StepEvaluator` 전략들을 호출합니다.
*   **시뮬레이션 흐름**: `simulate.go` (CLI) $\rightarrow$ `simulate_service.go` (오케스트레이션) $\rightarrow$ `sqlite_sandbox.go` (가상 데이터 생성).
*   **가상화 로직**: 샌드박스 모드일 때, `AnalyzeForecast`는 실제 PostgreSQL 어댑터 대신 `sqlite_virtual.go` 어댑터를 사용하여 운영 DB에 영향을 주지 않고 분석을 수행합니다.
