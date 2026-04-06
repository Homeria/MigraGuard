# ⚙️ MigraGuard v3.3 Implementation Details

본 문서는 v3.3에서 달성한 **정밀 데이터 수집 로직**과 **모듈화된 DB 레이어**의 상세 구현 명세를 다룹니다.

---

## 1. Modular DB Layer (`internal/db/`)

데이터 계층의 응집도를 높이고 결합도를 낮추기 위해 다음과 같이 모듈화되었습니다.

- **`models.go`**: `WorkloadSnapshot`, `TableDynamicMetrics` 등 시스템 전반에서 사용하는 도메인 모델 정의.
- **`postgres_adapter.go`**: PostgreSQL에서 누적 통계치와 실시간 테이블 지표(Size, Lag 등)를 읽어오는 전용 어댑터.
- **`sqlite_adapter.go`**: SQLite 연결 관리, 스키마 초기화, 데이터 보존 정책(Purge) 등 기초 인프라 관리.
- **`sqlite_repository.go`**: 수집기 전용. 델타 스냅샷 기록 및 차이값 계산을 위한 원본(Original) 데이터 동기화 담당.
- **`sqlite_analyzer.go`**: 분석기 전용. 리스크 엔진이 필요로 하는 통계 분석 쿼리(Baseline, TPS 등) 수행.

## 2. Background Collector Delta Engine (`internal/db/collector.go`)

- **Objective:** `pg_stat_statements`의 누적치로부터 수집 주기 사이의 실제 부하량(Delta)을 정밀하게 추출.
- **Delta Computation Logic:**
  1. **Fetch:** PostgreSQL로부터 현재 누적 스냅샷(`Current`)을 가져옴.
  2. **Lookup:** SQLite의 `original_pg_stat_statements`에서 이전 수집 시점의 누적치(`Previous`)를 조회.
  3. **Compute:** $\Delta = Current - Previous$.
  4. **Exception Handling:** 만약 $\Delta < 0$ (DB 통계 초기화 시)이라면 현재값(`Current`)을 그대로 델타로 채택.
  5. **Synchronize:** 다음 주기를 위해 현재값(`Current`)을 `original_pg_stat_statements`에 UPSERT.
  6. **Record:** 최종 계산된 $\Delta$를 `workload_snapshots`에 시계열 데이터로 저장.

## 3. High-Fidelity Service Simulation (`internal/simulation/` - v3.4 예정)

- **Objective:** `pgbench`의 한계를 넘어 실제 비즈니스 트래픽과 유사한 환경에서 리스크 엔진을 검증.
- **Simulation Strategy:**
  - **Worker Pool Architecture**: 여러 워커가 비동기적으로 DB 요청 수행.
  - **Query Mix Profile**: SELECT 80%, INSERT 15%, UPDATE 5% 등 비즈니스 시나리오별 쿼리 비율 설정.
  - **Dynamic TPS Control**: 설정된 프로파일에 따라 트래픽을 선형 또는 폭발적으로 증가시켜 리스크 점수 변화 유도.

## 4. Risk Evaluation Engine (`internal/engine/risk.go`)

- **Weighted Lambda ($\lambda_{final}$):**
  - $\lambda_{final} = \max(\lambda_{curr}, \lambda_{avg\_1h} \times 1.2, \lambda_{peak\_24h} \times 0.8)$.
  - 구간별 델타 데이터를 기반으로 하므로 $\lambda_{curr}$의 신뢰도가 v3.2 대비 비약적으로 상승.
- **Gatekeeping Logic:**
  - 분석 결과 `RiskScore >= 90%` 또는 `PermanentFailure` 감지 시 즉시 `Danger` 등급 부여 및 파이프라인 차단 시그널 송출.

---

## Technical Summary: The Refined Data Journey
1. **Agent**가 **Postgres**에서 누적치를 읽어 **SQLite**의 이전 상태와 비교 후 **Delta**만 저장.
2. **Analyze**가 SQL 파일을 파싱하여 변경 대상을 식별.
3. **Risk Engine**이 **SQLite**의 델타 이력과 PostgreSQL의 실시간 지표를 결합하여 최종 점수 산출.
4. **Load Generator**가 다양한 부하 상황을 연출하여 위 과정의 실효성을 반복 검증.

---
*Last Updated: 2026-04-06 (v3.3 Update)*
