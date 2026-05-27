# 🛡️ MigraGuard 아키텍처 리팩토링 계획서

본 문서는 `feat/adaptive-recommendation` (적응형 임계값 자동 제안 엔진) 구축 등 향후 고도화 작업을 앞두고, 시스템의 유지보수성 향상과 관심사 분리(SoC)를 위해 수립된 선제적 리팩토링 계획입니다. 

본 리팩토링은 현재의 '시뮬레이션 심화 고도화' 브랜치 작업을 완료하여 `develop` 브랜치로 무사히 통합한 이후, **독립된 전용 리팩토링 브랜치를 분기하여** 안전하게 수행할 예정입니다.

---

## 🎯 핵심 목표
1. **단일 책임 원칙(SRP) 준수:** 하나의 소스 파일이 하나의 역할과 책임만 가지도록 결합도를 낮춥니다.
2. **수학적 알고리즘과 하드웨어 인프라 격리:** 통계 연산 로직과 데이터베이스 I/O 계층을 분리하여 테스트 독립성을 확보합니다.
3. **가독성 및 유지보수 극대화:** 5단계 예측 모델의 확장성을 위해 개별 파일 단위로 모듈화합니다.

---

## 📂 1. 핵심 데이터 모델 분할 (`pkg/migraguard/types/models.go`)
현재 다양한 컨텍스트의 데이터 구조가 단일 파일에 몰려 있어, 향후 데이터 보고 모델 추가 시 유지보수 한계가 올 수 있습니다.

### [현황 및 문제점]
- DB 로우 메트릭, DDL 정적 분석 정보, 24시간 예측 데이터, 시뮬레이션용 YAML 구조체가 약 180줄에 걸쳐 누적 선언되어 있습니다.

### [개선 방향]
역할과 사용 주기에 맞춰 4개의 파일로 세분화합니다.
- **`models_core.go`:** `WorkloadSnapshot`, `TableDynamicMetrics`, `BaselineStats` (프로덕션 및 라이브 DB 메트릭 도메인 모델)
- **`models_analysis.go`:** `AnalysisResult`, `RiskConstants`, `RiskAnalysisReport`, `AnalysisResponse` (정량 리스크 분석 프로세스용 모델)
- **`models_forecast.go`:** `ForecastTimeSlot`, `ForecastReport` (24시간 가상 예보 데이터 모델)
- **`models_simulation.go`:** `SimulationScenario`, `VirtualPGState`, `SQLiteHistoryProfile`, `TimelineEvent` (Sandbox 시뮬레이션 전용 설정 모델)

---

## 📂 2. Sandbox 통계 알고리즘과 DB 드라이버 격리 (`pkg/migraguard/internal/infra/sqlite/sqlite_sandbox.go`)
트래픽 파형을 수학적으로 조율하는 부분과 SQLite에 쓰기 연산을 가하는 물리 드라이버가 혼재되어 있습니다.

### [현황 및 문제점]
- 피크 시프트, 비대칭 시간 왜곡(Time Warping), 노이즈 스케일링을 연산하는 `DefaultWorkloadProfiler` 구조체와, 이를 통해 계산된 데이터를 SQLite 테이블에 인서트하는 `SandboxEngine` 및 `SeedScenario`가 결합되어 있어 통계 알고리즘만의 단독 유닛 테스트가 어렵습니다.

### [개선 방향]
알고리즘 연산 계층과 데이터 액세스 계층을 독립된 소스 파일로 쪼갭니다.
- **`sandbox_profiler.go`:** `WorkloadProfiler` 인터페이스 및 `DefaultWorkloadProfiler` 통계 엔진의 수학적 공식을 여기로 격리합니다. (추후 적응형 추천 관련 데이터 피팅 로직도 이 모듈과 느슨하게 결합될 수 있어 확장성이 확보됩니다)
- **`sqlite_sandbox.go`:** `SandboxEngine`은 주입받은 `WorkloadProfiler`를 이용해 SQLite 세션 개설 및 대량 벌크 인서트 SQL 실행에만 고스란히 집중합니다.

---

## 📂 3. 5단계 리스크 평가 전략 개별 클래스화 (`pkg/migraguard/internal/analyzer/risk_evaluator.go`)
하나의 파일에 5개의 핵심 리스크 평가 로직이 직렬로 코딩되어 있습니다.

### [현황 및 문제점]
- `StepEvaluator` 인터페이스를 상속받은 5가지 평가기(`DDLTimeEvaluator`, `BlockingTimeEvaluator`, `PeakConnectionEvaluator`, `RecoveryTimeEvaluator`, `RiskScoreEvaluator`)가 병렬 나열되어 있습니다. 
- 추후 큐 대기 행렬 정교화나 하드웨어 병목 연산 수식이 심화될 시, 해당 파일의 소스 라인이 비정상적으로 증가하게 됩니다.

### [개선 방향]
각 단계별 평가기를 개별 소스 파일로 쪼개어 가독성을 높입니다.
- **`risk_evaluator.go`:** `StepEvaluator` 공통 인터페이스만 유지합니다.
- **`evaluator_01_ddl_time.go`:** 디스크 I/O 속도 대비 테이블 재작성 및 DDL 수행 시간 예측 ($T_{ddl}$).
- **`evaluator_02_blocking.go`:** DDL 락 영향도 가중치를 반영한 라이브 트래픽 대기 시간 예측 ($T_{block}$).
- **`evaluator_03_connections.go`:** 단위 시간당 TPS와 대기 지연에 따른 누적 연결 수 예측 ($C_{peak}$).
- **`evaluator_04_recovery.go`:** 과부하 커넥션 배출 시간 및 영구 장애 판정 계산 ($T_{rec}$).
- **`evaluator_05_scoring.go`:** 시스템 가용한도 대비 리스크 스코어 매핑 및 최종 경고 단계 결정.

---

## 🚀 기대 효과 및 로드맵 연계
- **SoC (관심사 분리):** 개발자는 데이터베이스 입출력 버그를 두려워하지 않고 순수 통계 프로파일 공식을 안전하게 확장 및 테스트할 수 있습니다.
- **Feat/Adaptive-Recommendation 시너지:** 통계 분석 및 임계점 피팅 로직을 추가해야 할 때, 이미 분리된 `sandbox_profiler.go` 에 통계 함수를 깨끗하게 증설할 수 있으며 도메인 모델(`models_analysis.go`)의 구조 역시 타 레이어 영향 없이 증설 가능해집니다.
