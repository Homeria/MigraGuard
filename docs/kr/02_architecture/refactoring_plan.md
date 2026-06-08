# 🛡️ MigraGuard 아키텍처 리팩토링 계획 및 결과 보고서

본 문서는 `feat/adaptive-recommendation` (적응형 임계값 자동 제안 엔진) 구축 등 향후 고도화 작업을 앞두고, 시스템의 유지보수성 향상과 관심사 분리(SoC)를 위해 수립된 선제적 리팩토링 계획과 그에 따른 최종 완수 결과를 공식 기록한 문서입니다.

> [!NOTE]
> 본 아키텍처 리팩토링 계획서에 기재된 모든 계획은 **2026-05-27 부로 100% 완벽히 수행 및 검증 완료**되었습니다.

---

## 🎯 핵심 목표 및 달성도
1. **단일 책임 원칙(SRP) 준수:** 하나의 소스 파일이 하나의 역할과 책임만 가지도록 결합도를 낮춤 **[완료]**
2. **수학적 알고리즘과 하드웨어 인프라 격리:** 통계 연산 로직과 데이터베이스 I/O 계층을 분리하여 테스트 독립성을 확보함 **[완료]**
3. **가독성 및 유지보수 극대화:** 5단계 예측 모델의 확장성을 위해 개별 파일 단위로 모듈화함 **[완료]**
4. **결함 없는 예외 전파 및 동시성 격리:** 에러 묵살 제거, `rows.Err()` 검사 강제화, 스레드-로컬 난수 인스턴스화를 통한 안전성 고도화 **[신규 추가 및 완료]**

---

## 📂 1. 핵심 데이터 모델 분할 (`pkg/migraguard/types/`)
다양한 컨텍스트의 데이터 구조가 단일 파일에 몰려 있던 문제를 해결하기 위해 역할에 맞춰 4개의 파일로 분할을 완료했습니다.
- **`models_core.go`:** `WorkloadSnapshot`, `TableDynamicMetrics`, `BaselineStats` (프로덕션 및 라이브 DB 메트릭 도메인 모델)
- **`models_analysis.go`:** `AnalysisResult`, `RiskConstants`, `RiskAnalysisReport`, `AnalysisResponse` (정량 리스크 분석 프로세스용 모델)
- **`models_forecast.go`:** `ForecastTimeSlot`, `ForecastReport` (24시간 가상 예보 데이터 모델)
- **`models_simulation.go`:** `SimulationScenario`, `VirtualPGState`, `SQLiteHistoryProfile`, `TimelineEvent` (Sandbox 시뮬레이션 전용 설정 모델)

---

## 📂 2. Sandbox 통계 알고리즘과 DB 드라이버 격리 (`pkg/migraguard/internal/infra/sqlite/`)
트래픽 파형을 수학적으로 조율하는 통계 연산 계층과 데이터 액세스 계층을 독립된 소스 파일로 쪼갰습니다.
- **`sandbox_profiler.go`:** `WorkloadProfiler` 인터페이스 및 `DefaultWorkloadProfiler` 통계 엔진의 수학적 공식을 여기로 완전히 격리했습니다.
- **`sqlite_sandbox.go`:** `SandboxEngine`은 주입받은 `WorkloadProfiler`를 이용해 SQLite 세션 개설 및 대량 벌크 인서트 SQL 실행에만 고스란히 집중하게 단순화되었습니다.

---

## 📂 3. 5단계 리스크 평가 전략 개별 패키지화 (`pkg/migraguard/internal/analyzer/`)
하나의 파일에 5개의 핵심 리스크 평가 로직이 직렬로 코딩되어 있던 것을 별도의 전용 서브 패키지(`evaluators`) 하위의 개별 소스로 온전히 아토믹 분리했습니다.
- **`evaluator.go`:** `StepEvaluator` 공통 인터페이스만 명시합니다.
- **`ddl_time.go`:** 디스크 I/O 속도 대비 테이블 재작성 및 DDL 수행 시간 예측 ($T_{ddl}$).
- **`blocking.go`:** DDL 락 영향도 가중치를 반영한 라이브 트래픽 대기 시간 예측 ($T_{block}$).
- **`connections.go`:** 단위 시간당 TPS와 대기 지연에 따른 누적 연결 수 예측 ($C_{peak}$).
- **`recovery.go`:** 과부하 커넥션 배출 시간 및 영구 장애 판정 계산 ($T_{rec}$).
- **`scoring.go`:** 시스템 가용한도 대비 리스크 스코어 매핑 및 최종 경고 단계 결정.

---

## 📂 4. [신규 완수] 4대 예외 안전성 및 로깅 구조화 고도화
v3.9 시뮬레이터 통합의 완성도를 올리기 위해 다음의 4가지 아키텍처 정교화 작업을 안전하게 완수했습니다.
- **`simulate_service.go` 로깅 인터페이스 통합**: 날것의 콘솔 표준 I/O(`fmt`) 호출을 차단하고, 주입받은 `types.Logger` 인터페이스로 통합 제어하도록 개편하여 로깅 관심사 격리를 완료했습니다.
- **데이터베이스 `rows.Err()` 의무 검증**: SQLite 및 PostgreSQL 데이터 수집용 `rows.Next()` 루프 직후 `rows.Err()`을 반드시 확인하여 트랜잭션 도중의 데이터 유실/단절을 예방하고, `Scan` 오류 발생 시 에러를 정상 에스컬레이션하도록 전파력을 강화했습니다.
- **예측 분석기 SQLite 에러 묵살 제거**: `risk_calculator.go` 내에서 통계 지표를 가져올 때 발생하는 SQLite 쿼리 에러 묵살 처리(`_`)를 배제하고 명시적 랩핑 가드를 세워 예측 오류의 투명성을 확보했습니다.
- **샌드박스 Prepare 및 동시성 난수 격리**: `tx.Prepare` 오류의 가드를 확고히 하고 리소스 Close 처리를 강화하였으며, 글로벌 난수(`rand.Intn`)를 스레드-세이프한 로컬 난수 인스턴스 `rng := rand.New(rand.NewSource(...))`로 격리하여 멀티스레드 경합 요소를 제거했습니다.

---

## 🚀 아키텍처적 기대 효과
- **SoC (관심사 분리):** 개발자는 데이터베이스 입출력 버그를 두려워하지 않고 순수 통계 프로파일 공식을 안전하게 확장 및 테스트할 수 있습니다.
- **Feat/Adaptive-Recommendation 시너지:** 통계 분석 및 임계점 피팅 로직을 추가해야 할 때, 이미 분리된 `sandbox_profiler.go` 에 통계 함수를 깨끗하게 증설할 수 있으며 도메인 모델(`models_analysis.go`)의 구조 역시 타 레이어 영향 없이 증설 가능해집니다.
