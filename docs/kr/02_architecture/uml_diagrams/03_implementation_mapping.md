# Level 3: 구현 추적성 매트릭스 (Traceability Matrix)

이 문서는 Level 0-2에서 설명된 아키텍처 도메인 및 워크플로우를 실제 Go 소스 코드 및 함수와 매핑합니다.

---

## 1. Command 도메인 (인터페이스 레이어)
*Cobra CLI를 통한 사용자 상호작용의 진입점입니다.*

| 워크플로우 | 진입점 (Go 파일) | 주요 함수/메서드 | 설명 |
| :--- | :--- | :--- | :--- |
| `analyze` | `cmd/migraguard/analyze.go` | `runAnalyze()` | 분석 명령어의 기본 오케스트레이터입니다. |
| `simulate` | `cmd/migraguard/simulate.go` | `runSimulate()` | 시나리오 로드 및 샌드박스 초기화를 처리합니다. |
| `check` (Agent) | `cmd/migraguard/check.go` | `runCheck()` | 백그라운드 메트릭 수집 에이전트를 시작합니다. |

---

## 2. Application 도메인 (서비스 레이어)
*CLI와 Core SDK를 연결하는 오케스트레이션 로직입니다.*

| 서비스 | 소스 파일 | 핵심 로직 메서드 | 역할 |
| :--- | :--- | :--- | :--- |
| **Analyze Service** | `pkg/migraguard/internal/app/analyze_service.go` | `Run()` | SQL 파싱, 리스크 평가, 예측 CSV 내보내기를 조정합니다. |
| **Simulate Service** | `pkg/migraguard/internal/app/simulate_service.go` | `Run()` | 시나리오 YAML 언마샬링 및 샌드박스 시딩을 관리합니다. |
| **Agent Service** | `pkg/migraguard/internal/app/agent_service.go` | `Run()` | 주기적인 메트릭 수집을 위한 무한 루프를 구현합니다. |

---

## 3. Domain Logic: 리스크 엔진 (전략 레이어)
*5단계 모델을 구현하는 시스템의 수학적 핵심입니다.*

| 아키텍처 컴포넌트 | 소스 파일 | 주요 심볼/함수 | 설명 |
| :--- | :--- | :--- | :--- |
| **Orchestrator** | `pkg/migraguard/internal/analyzer/risk_calculator.go` | `AnalyzeRisk()` | 단일 시점에 대해 StepEvaluator 시퀀스를 실행합니다. |
| **Predictive Engine** | `pkg/migraguard/internal/analyzer/risk_calculator.go` | `AnalyzeForecast()` | **TPS Tie-Breaker** 로직을 사용하여 24시간 가상 시뮬레이션을 실행합니다. |
| **SQL Parser** | `pkg/migraguard/internal/analyzer/sql_parser.go` | `ParseSQL()` | `pg_query_go`를 사용하여 SQL을 AST로 변환하고 DDL 유형을 식별합니다. |
| **Evaluators (Steps 1-5)** | `pkg/migraguard/internal/analyzer/risk_evaluator.go` | `StepEvaluator` (인터페이스) | $T_{ddl}, T_{block}, C_{peak}, T_{rec}$, 최종 점수를 위한 전략 패턴 구현체입니다. |

---

## 4. Infrastructure 도메인 (어댑터 레이어)
*데이터베이스와 상호작용하기 위한 인터페이스입니다.*

| 어댑터 | 소스 파일 | 주요 책임 | 구현 세부사항 |
| :--- | :--- | :--- | :--- |
| **PostgreSQL Adapter** | `pkg/migraguard/internal/infra/postgres/adapter.go` | `FetchTableDynamicMetrics()` | `pg_stat_statements` 및 시스템 카탈로그에서 실시간 메트릭을 추출합니다. |
| **SQLite Repository** | `pkg/migraguard/internal/infra/sqlite/sqlite_repository.go` | `InitializeSchema()` | 로컬 메트릭 저장소 및 샌드박스 파일을 관리합니다. |
| **Virtual PG Adapter** | `pkg/migraguard/internal/infra/sqlite/sqlite_virtual.go` | `FetchTableDynamicMetrics()` | **시뮬레이션 핵심**: SQLite 샌드박스 데이터를 읽어 PostgreSQL 동작을 모방합니다. |
| **Sandbox Engine** | `pkg/migraguard/internal/infra/sqlite/sqlite_sandbox.go` | `SeedScenario()` | Sine 파형 및 노이즈 분산을 기반으로 시계열 데이터를 생성합니다. |

---

## 5. 시각화 파이프라인 (Visualization Pipeline)
*Go 분석과 Python 리포팅 간의 브릿지입니다.*

| 컴포넌트 | 소스 파일 | 트리거 지점 | 설명 |
| :--- | :--- | :--- | :--- |
| **CSV Exporter** | `pkg/migraguard/internal/app/analyze_service.go` | `exportForecastCSV()` | 분석 성공 후 `predictive_forecast.csv`를 생성합니다. |
| **Python Plotter** | `tools/visualization/analyze/plot_predictive_heatmap.py` | `plot_predictive_heatmap()` | **이중 Y축 히트맵**(TPS & P99)을 Matplotlib으로 구현합니다. |
| **Orchestrator Script**| `scripts/cmd/analyze/run-predictive-forecast.bat` | 해당 없음 | Go CLI와 Python Plotter 실행을 연결하는 배치/쉘 스크립트입니다. |
