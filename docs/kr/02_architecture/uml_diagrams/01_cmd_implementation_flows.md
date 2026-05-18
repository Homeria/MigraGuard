# Level 1: CLI 명령어 구현 흐름 (Implementation Flows)

이 문서는 MigraGuard가 지원하는 모든 CLI 명령어의 실행 경로를 함수 단위의 상세 노드로 시각화합니다. 각 노드는 실제 구현 파일과 함수 정보를 포함합니다.

---

## 1. `analyze` 명령어 (실시간 분석 / 샌드박스 / 예측)
사용자의 DDL 위험도를 분석하고 미래 트래픽을 예측하는 핵심 명령어입니다.

```mermaid
flowchart TD
    A1["<b>File:</b> root.go / analyze.go<br/><b>Func:</b> runAnalyze()<br/><b>Args:</b> args, flags<br/><b>Role:</b> CLI 입력값 검증 및 AnalyzeService 호출"]
    
    A2["<b>File:</b> analyze_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> task (SQLPath, Forecast)<br/><b>Role:</b> 분석 파이프라인 전체 제어"]
    
    A3["<b>File:</b> sql_parser.go<br/><b>Func:</b> ParseSQL()<br/><b>Args:</b> sqlContent<br/><b>Role:</b> SQL AST 분석 및 DDL 메타데이터 추출"]
    
    subgraph Execution_Path [분석 전략 선택]
        A4_Live["<b>File:</b> postgres/adapter.go<br/><b>Func:</b> FetchTableDynamicMetrics()<br/><b>Args:</b> tableName<br/><b>Role:</b> 실재 DB에서 실시간 지표 추출"]
        
        A4_Sandbox["<b>File:</b> sqlite_virtual.go<br/><b>Func:</b> FetchTableDynamicMetrics()<br/><b>Args:</b> tableName<br/><b>Role:</b> SQLite 샌드박스에서 가상 지표 추출"]
    end
    
    A5["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeRisk()<br/><b>Args:</b> ctx, analysisResult<br/><b>Role:</b> 5단계 정량적 리스크 모델 실행"]
    
    A6["<b>File:</b> risk_calculator.go<br/><b>Func:</b> AnalyzeForecast()<br/><b>Args:</b> ctx, forecastData<br/><b>Role:</b> 24시간 가상 시뮬레이션 루프 (Tie-breaker 적용)"]
    
    A7["<b>File:</b> analyze_service.go<br/><b>Func:</b> exportForecastCSV()<br/><b>Args:</b> forecastReports<br/><b>Role:</b> 시각화용 CSV 생성"]

    A1 --> A2
    A2 --> A3
    A3 --> A4_Live
    A3 --> A4_Sandbox
    A4_Live & A4_Sandbox --> A5
    A5 -->|--forecast 활성화 시| A6
    A6 --> A7
    
    style A5 fill:#fff3e0,stroke:#e65100,stroke-width:2px
    style A6 fill:#fff3e0,stroke:#e65100,stroke-width:2px
```

---

## 2. `simulate` 명령어 (샌드박스 생성)
가상의 트래픽 시나리오를 생성하여 연구 및 실험 환경을 구축합니다.

```mermaid
flowchart TD
    S1["<b>File:</b> simulate.go<br/><b>Func:</b> runSimulate()<br/><b>Args:</b> scenarioPath, force<br/><b>Role:</b> 시나리오 파일 경로 확인 및 SimulateService 호출"]
    
    S2["<b>File:</b> simulate_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> scenarioPath, force<br/><b>Role:</b> YAML 파싱 및 샌드박스 DB 파일(.db) 관리"]
    
    S3["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> SeedScenario()<br/><b>Args:</b> scenario<br/><b>Role:</b> 7일치 가상 데이터 생성 트랜잭션 관리"]
    
    S4["<b>File:</b> sqlite_sandbox.go<br/><b>Func:</b> CalculateTPS()<br/><b>Args:</b> time, profile<br/><b>Role:</b> 수학적 모델(Sine/Weekly/Events) 기반 TPS 산출"]

    S1 --> S2
    S2 --> S3
    S3 --> S4
    
    style S3 fill:#f1f8e9,stroke:#33691e,stroke-width:2px
```

---

## 3. `check` 명령어 (백그라운드 에이전트)
운영 DB의 메트릭을 지속적으로 수집하여 SQLite에 저장합니다.

```mermaid
flowchart TD
    C1["<b>File:</b> check.go<br/><b>Func:</b> runCheck()<br/><b>Args:</b> targetTables<br/><b>Role:</b> 에이전트 서비스 설정 및 실행"]
    
    C2["<b>File:</b> agent_service.go<br/><b>Func:</b> Run()<br/><b>Args:</b> ctx<br/><b>Role:</b> 지정된 인터벌(1m) 마다 반복 실행되는 무한 루프"]
    
    C3["<b>File:</b> postgres/adapter.go<br/><b>Func:</b> FetchCurrentWorkloadSnapshot()<br/><b>Args:</b> ctx<br/><b>Role:</b> pg_stat_statements 뷰에서 쿼리 통계 수집"]
    
    C4["<b>File:</b> sqlite_repository.go<br/><b>Func:</b> UpdateWorkloadSnapshots()<br/><b>Args:</b> snapshots<br/><b>Role:</b> 수집된 데이터를 SQLite에 저장 (UPSERT)"]
    
    C5["<b>File:</b> sqlite_init.go<br/><b>Func:</b> MaintenancePurgeData()<br/><b>Args:</b> retentionDays<br/><b>Role:</b> 보관 기간이 지난 오래된 데이터 삭제"]

    C1 --> C2
    C2 --> C3
    C3 --> C4
    C4 --> C5
    C5 -->|Loop| C2
    
    style C2 fill:#e1f5fe,stroke:#01579b,stroke-width:2px
```

---

## 4. `export` 명령어 (데이터 추출)
수집되거나 생성된 샌드박스 데이터를 연구용 CSV로 추출합니다.

```mermaid
flowchart TD
    E1["<b>File:</b> export.go<br/><b>Func:</b> runExport()<br/><b>Args:</b> outputPath<br/><b>Role:</b> 출력 경로 확인 및 클라이언트 메서드 호출"]
    
    E2["<b>File:</b> client.go<br/><b>Func:</b> ExportSandboxMetrics()<br/><b>Args:</b> outputPath<br/><b>Role:</b> 파일 스트림 생성 및 익스포트 오케스트레이션"]
    
    E3["<b>File:</b> sqlite_repository.go<br/><b>Func:</b> FetchAllTableMetrics()<br/><b>Args:</b> ctx<br/><b>Role:</b> SQLite에서 보관된 모든 지표 데이터 쿼리"]
    
    E4["<b>File:</b> client.go<br/><b>Func:</b> Write CSV Rows<br/><b>Args:</b> metrics, writer<br/><b>Role:</b> 조회된 데이터를 CSV 형식으로 파일 저장"]

    E1 --> E2
    E2 --> E3
    E3 --> E4
    
    style E2 fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
```
