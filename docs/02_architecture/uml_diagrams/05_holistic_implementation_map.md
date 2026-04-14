# 🗺️ MigraGuard 계층형 전역 구현 지도 (Layered Full Detail Map)

본 문서는 MigraGuard의 **기능 중심 아키텍처(v3.6)**를 기반으로, 각 레이어별 파일 경로와 함수의 상세 시그니처를 기술한 통합 설계도입니다.

## 1. 설계 의도
- **구조 개편 반영**: `analyzer`, `collector`, `infra`, `app`, `shared`로 재편된 최신 디렉토리 구조를 완벽히 시각화함.
- **추적성 극대화**: 상위 어플리케이션 레이어부터 하위 인프라 레이어까지의 호출 경로를 파일 단위로 명시함.

## 2. 계층형 전역 구현 지도 (Mermaid)

```mermaid
graph TD
    %% ==========================================
    %% 1. CLI Layer
    %% ==========================================
    subgraph LAYER_CLI [<b>1. CLI & Entry Points Layer</b>]
        style LAYER_CLI fill:#f9f5ff,stroke:#7c3aed,stroke-width:2px

        subgraph CMD_ANALYZE ["<b>cmd/migraguard/analyze.go</b>"]
            CA_Run["<b>Run()</b><br/>Args: cmd, args<br/>Returns: error<br/>Desc: 분석 서비스 기동 및 게이트키핑 종료 코드 결정"]
        end

        subgraph CMD_AGENT ["<b>cmd/migraguard/agent.go</b>"]
            CG_Run["<b>Run()</b><br/>Args: cmd, args<br/>Returns: error<br/>Desc: OS 신호 감지 및 에이전트 서비스 백그라운드 가동"]
        end
    end

    %% ==========================================
    %% 2. App Layer (Application Services)
    %% ==========================================
    subgraph LAYER_APP [<b>2. Application Service Layer</b>]
        style LAYER_APP fill:#f0f9ff,stroke:#0284c7,stroke-width:2px

        subgraph APP_ANALYZE ["<b>internal/app/analyze.go</b>"]
            AA_Run["<b>Run()</b><br/>Args: ctx, task<br/>Returns: *AnalysisResponse, error<br/>Desc: 파싱/검증/분석 파이프라인 전체 조율"]
        end

        subgraph APP_AGENT ["<b>internal/app/agent.go</b>"]
            AG_Run["<b>Run()</b><br/>Args: ctx<br/>Returns: error<br/>Desc: Collector 초기화 및 라이프사이클 관리"]
        end
    end

    %% ==========================================
    %% 3. Domain Layer (The Logic)
    %% ==========================================
    subgraph LAYER_DOMAIN [<b>3. Core Domain Layer</b>]
        style LAYER_DOMAIN fill:#f0fdf4,stroke:#16a34a,stroke-width:2px

        subgraph DOMAIN_ANALYZER ["<b>internal/analyzer/</b>"]
            PA_Parse["<b>ast.go: ParseSQL()</b><br/>Args: sql: string<br/>Returns: []AnalysisResult, error<br/>Desc: SQL AST 분석 및 Rewrite 여부 판별"]
            ER_Analyze["<b>risk.go: AnalyzeRisk()</b><br/>Args: ctx, res<br/>Returns: *RiskAnalysisReport, error<br/>Desc: 5단계 수식 기반 정량적 리스크 산출"]
        end

        subgraph DOMAIN_COLLECTOR ["<b>internal/collector/</b>"]
            DC_Start["<b>worker.go: Start()</b><br/>Args: ctx<br/>Returns: (없음)<br/>Desc: 주기적 Ticker 루프 및 델타 계산 실행"]
            DC_Compute["<b>worker.go: computeDelta()</b><br/>Args: curr, prev<br/>Returns: []WorkloadSnapshot<br/>Desc: 누적치 차이를 통한 시점별 TPS 산출"]
        end
    end

    %% ==========================================
    %% 4. Infra Layer (The Adapters)
    %% ==========================================
    subgraph LAYER_INFRA [<b>4. Infrastructure & Persistence Layer</b>]
        style LAYER_INFRA fill:#fffbeb,stroke:#d97706,stroke-width:2px

        subgraph INFRA_POSTGRES ["<b>internal/infra/postgres/</b>"]
            DP_Metrics["<b>adapter.go: FetchTableDynamicMetrics()</b><br/>Args: ctx, name<br/>Returns: *TableDynamicMetrics, error<br/>Desc: 테이블별 실시간 상세 지표 수집"]
            DP_Schema["<b>adapter.go: CheckTableSchemaPresence()</b><br/>Args: ctx, name, cols<br/>Returns: error<br/>Desc: DB 스키마 존재 여부 사전 검증"]
        end

        subgraph INFRA_SQLITE ["<b>internal/infra/sqlite/</b>"]
            DS_Record["<b>sqlite_repository.go: RecordDeltaSnapshots()</b><br/>Args: snapshots<br/>Returns: error<br/>Desc: 계산된 델타 지표를 SQLite에 영속화"]
            DS_Base["<b>sqlite_analyzer.go: GetTableBaselineStatistics()</b><br/>Args: name<br/>Returns: *BaselineStats, error<br/>Desc: 과거 평균 및 피크 트래픽 통계 산출"]
            DS_Window["<b>sqlite_analyzer.go: IdentifySafestDeploymentWindow()</b><br/>Args: (없음)<br/>Returns: string, float64, error<br/>Desc: 최적의 배포 시간대 자동 추천"]
        end
    end

    %% ==========================================
    %% 5. Shared Layer (The Foundation)
    %% ==========================================
    subgraph LAYER_SHARED [<b>5. Shared Foundation Layer</b>]
        style LAYER_SHARED fill:#f8fafc,stroke:#64748b,stroke-width:1px

        subgraph SHARED_TYPES ["<b>internal/shared/types/</b>"]
            T_Ent["<b>models.go: Entity Definitions</b><br/>WorkloadSnapshot, RiskAnalysisReport 등"]
            T_Iface["<b>interfaces.go: Interface Definitions</b><br/>PostgresClient, SQLiteClient 인터페이스"]
        end
    end

    %% --- 호출 관계 (Cross-Layer Call Flow) ---
    CA_Run --> AA_Run
    AA_Run --> PA_Parse
    AA_Run --> DP_Schema
    AA_Run --> ER_Analyze
    
    ER_Analyze --> DP_Metrics
    ER_Analyze --> DS_Base
    ER_Analyze --> DS_Window

    CG_Run --> AG_Run
    AG_Run --> DC_Start
    DC_Start --> DC_Compute
    DC_Compute --> DS_Record
```

## 3. 기능 중심 레이어별 책임 요약

| 레이어 (Layer) | 주요 파일 위치 | 핵심 책임 |
| :--- | :--- | :--- |
| **CLI & Entry** | `cmd/` | 사용자 인터페이스 제공 및 OS 신호 처리. |
| **Application** | `internal/app` | 유스케이스 조율 및 도메인 로직 실행 흐름 제어. |
| **Domain** | `internal/analyzer`, `internal/collector` | 시스템의 핵심 비즈니스 로직 및 알고리즘 수행. |
| **Infrastructure** | `internal/infra` | 데이터베이스 연결 및 외부 라이브러리와의 저수준 상호작용. |
| **Shared** | `internal/shared` | 프로젝트 전역에서 사용하는 데이터 모델, 인터페이스, 에러 규격 보관. |

---
*Last Updated: 2026-04-14 (v3.6 Feature-First Architecture Finalized)*
