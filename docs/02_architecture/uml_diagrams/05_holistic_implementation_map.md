# 🗺️ MigraGuard 계층형 전역 구현 지도 (Layered Full Detail Map)

본 문서는 MigraGuard의 소스 코드 구조를 **5개의 논리적 계층(Layer)**으로 분류하고, 각 계층 내 파일(.go)과 함수의 상세 시그니처 및 역할을 기술한 통합 설계도입니다.

## 1. 설계 의도
- **계층형 아키텍처 시각화**: CLI -> Service -> Core -> Infra로 이어지는 계층 구조를 명확히 함.
- **정보의 밀집 및 조직화**: 파일 단위의 상세 정보를 유지하면서도, 대단위 컴포넌트 간의 관계를 체계화함.
- **공학적 완성도**: 캡스톤 디자인 심사 시 시스템의 모듈화 및 계층 분리 설계를 증명하는 용도로 최적화함.

## 2. 계층형 전역 구현 지도 (Mermaid)

```mermaid
graph TD
    %% ==========================================
    %% 1. CLI Layer (사용자 접점)
    %% ==========================================
    subgraph LAYER_CLI [<b>1. CLI & Entry Points Layer</b>]
        style LAYER_CLI fill:#f9f5ff,stroke:#7c3aed,stroke-width:2px

        subgraph CMD_ANALYZE ["<b>cmd/migraguard/analyze.go</b>"]
            CA_Run["<b>Run()</b><br/>Args: cmd, args<br/>Returns: error<br/>Desc: 사용자 SQL 경로 수신 및 분석 서비스 기동"]
        end

        subgraph CMD_AGENT ["<b>cmd/migraguard/agent.go</b>"]
            CG_Run["<b>Run()</b><br/>Args: cmd, args<br/>Returns: error<br/>Desc: 수집 주기 설정 및 에이전트 서비스 기동"]
        end

        subgraph CMD_LOADGEN ["<b>cmd/loadgen/main.go</b>"]
            CL_Main["<b>main()</b><br/>Args: (없음)<br/>Returns: (없음)<br/>Desc: 부하 시뮬레이션 워커 및 시나리오 시작"]
        end
    end

    %% ==========================================
    %% 2. Service Layer (비즈니스 오케스트레이션)
    %% ==========================================
    subgraph LAYER_SERVICE [<b>2. Domain Service Layer</b>]
        style LAYER_SERVICE fill:#f0f9ff,stroke:#0284c7,stroke-width:2px

        subgraph SVC_ANALYZE ["<b>internal/service/analyze.go</b>"]
            SA_Run["<b>Run()</b><br/>Args: ctx, task: AnalysisTask<br/>Returns: *AnalysisResponse, error<br/>Desc: 파싱/검증/리스크 분석의 전체 흐름 제어"]
        end

        subgraph SVC_AGENT ["<b>internal/service/agent.go</b>"]
            SG_Run["<b>Run()</b><br/>Args: ctx<br/>Returns: error<br/>Desc: 데이터 수집기(Collector) 생성 및 루프 관리"]
        end
    end

    %% ==========================================
    %% 3. Core Engine Layer (핵심 알고리즘 및 파서)
    %% ==========================================
    subgraph LAYER_CORE [<b>3. Core Engine & Parser Layer</b>]
        style LAYER_CORE fill:#f0fdf4,stroke:#16a34a,stroke-width:2px

        subgraph PARSER_AST ["<b>internal/parser/ast.go</b>"]
            PA_Parse["<b>ParseSQL()</b><br/>Args: sql: string<br/>Returns: []AnalysisResult, error<br/>Desc: SQL AST 분석 및 Rewrite 여부 판별"]
        end

        subgraph ENGINE_RISK ["<b>internal/engine/risk.go</b>"]
            ER_Analyze["<b>AnalyzeRisk()</b><br/>Args: ctx, res: AnalysisResult<br/>Returns: *RiskAnalysisReport, error<br/>Desc: 5단계 리스크 수식 기반 정량적 분석 수행"]
        end

        subgraph ENGINE_RULES ["<b>internal/engine/rules.go</b>"]
            EU_Eval["<b>EvaluateLevel()</b><br/>Args: score: float64<br/>Returns: level: string<br/>Desc: 점수에 따른 Safe/Warning/Danger 판정"]
        end
    end

    %% ==========================================
    %% 4. Infrastructure Layer (데이터 저장 및 수집)
    %% ==========================================
    subgraph LAYER_INFRA [<b>4. Infrastructure & Persistence Layer</b>]
        style LAYER_INFRA fill:#fffbeb,stroke:#d97706,stroke-width:2px

        subgraph DB_COLLECTOR ["<b>internal/db/collector.go</b>"]
            DC_Start["<b>Start()</b><br/>Args: ctx<br/>Returns: (없음)<br/>Desc: 백그라운드 지표 수집 및 델타 계산 실행"]
            DC_Compute["<b>computeDelta()</b><br/>Args: curr, prev<br/>Returns: []WorkloadSnapshot<br/>Desc: 누적치 차이를 통한 시점별 TPS 산출"]
        end

        subgraph DB_POSTGRES ["<b>internal/db/postgres_adapter.go</b>"]
            DP_Snapshot["<b>FetchCurrentWorkloadSnapshot()</b><br/>Args: ctx<br/>Returns: []WorkloadSnapshot, error<br/>Desc: 운영 DB에서 원시 누적 통계치 조회"]
            DP_Metrics["<b>FetchTableDynamicMetrics()</b><br/>Args: ctx, name<br/>Returns: *TableDynamicMetrics, error<br/>Desc: 테이블별 실시간 상세 지표 수집"]
            DP_Schema["<b>CheckTableSchemaPresence()</b><br/>Args: ctx, name, cols<br/>Returns: error<br/>Desc: DB 스키마 존재 여부 사전 검증"]
        end

        subgraph DB_SQLITE ["<b>internal/db/sqlite_adapter.go</b>"]
            DS_Record["<b>RecordDeltaSnapshots()</b><br/>Args: snapshots<br/>Returns: error<br/>Desc: 계산된 델타 지표를 SQLite에 영속화"]
            DS_TPS["<b>GetRecentTPSByDelta()</b><br/>Args: tableName<br/>Returns: float64, error<br/>Desc: SQLite에서 최근 실시간 TPS 정보 로드"]
            DS_Base["<b>GetTableBaselineStatistics()</b><br/>Args: tableName<br/>Returns: *BaselineStats, error<br/>Desc: 과거 평균 및 피크 트래픽 통계 산출"]
            DS_Window["<b>IdentifySafestDeploymentWindow()</b><br/>Args: (없음)<br/>Returns: string, float64, error<br/>Desc: 최적의 배포 시간대 자동 추천"]
        end
    end

    %% ==========================================
    %% 5. Simulation Layer (검증 도구)
    %% ==========================================
    subgraph LAYER_SIM [<b>5. Simulation & Validation Layer</b>]
        style LAYER_SIM fill:#fef2f2,stroke:#dc2626,stroke-width:1px

        subgraph SIM_ENGINE ["<b>internal/simulation/engine.go</b>"]
            SE_Run["<b>RunSimulation()</b><br/>Args: ctx, cfg<br/>Returns: error<br/>Desc: 워커 풀 생성 및 부하 생성 실행"]
            SE_Sine["<b>SineWaveIntensity()</b><br/>Args: hour: int<br/>Returns: float64<br/>Desc: 시간대별 유동적 부하 가중치 계산"]
        end
    end

    %% --- 계층 간 호출 관계 (Cross-Layer Call Flow) ---
    CA_Run --> SA_Run
    SA_Run --> PA_Parse
    SA_Run --> DP_Schema
    SA_Run --> ER_Analyze
    
    ER_Analyze --> DP_Metrics
    ER_Analyze --> DS_TPS
    ER_Analyze --> DS_Base
    ER_Analyze --> DS_Window
    ER_Analyze --> EU_Eval

    CG_Run --> SG_Run
    SG_Run --> DC_Start
    DC_Start --> DP_Snapshot
    DC_Start --> DC_Compute
    DC_Compute --> DS_Record
    
    CL_Main --> SE_Run
    SE_Run --> SE_Sine
```

## 3. 계층별 책임 및 데이터 흐름 요약

| 계층 (Layer) | 주요 역할 | 핵심 데이터 흐름 |
| :--- | :--- | :--- |
| **CLI Layer** | 사용자 인터페이스 제공 및 명령 전달 | CLI Flag/Args -> Service Task |
| **Service Layer** | 비즈니스 로직 오케스트레이션 | Task -> Engine 호출 -> Report 생성 |
| **Core Layer** | SQL 분석 및 리스크 산출 알고리즘 수행 | SQL String -> AST -> Risk Score |
| **Infra Layer** | 데이터베이스 접근 및 시계열 지표 관리 | Postgres Stat -> Delta Calculation -> SQLite |
| **Simulation Layer** | 시스템 검증을 위한 인위적 부하 발생 | Time Config -> Sine Curve -> DB Transaction |

---
*Last Updated: 2026-04-14 (Layered Full Detail Map Finalized)*
