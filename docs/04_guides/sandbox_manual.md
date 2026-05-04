# 🏜️ MigraGuard 시뮬레이션 샌드박스 운용 가이드 (Sandbox Manual)

본 문서는 실제 PostgreSQL 데이터베이스 연결 없이도, 선언적인 시나리오 정의를 통해 가상의 부하 환경을 구축하고 마이그레이션 리스크를 검증하는 **Simulation Sandbox** 기능의 상세 사용법을 안내합니다.

---

## 1. 개요 (Overview)

MigraGuard의 시뮬레이션 샌드박스는 연구 및 검증의 재현성을 위해 설계되었습니다. 실제 운영 환경에서 7일간 데이터를 수집할 필요 없이, 수학적 모델을 기반으로 생성된 7일~30일치 시계열 데이터를 1초 만에 SQLite 파일로 생성합니다.

### 핵심 가치
*   **오프라인 검증**: 인터넷이나 DB 연결이 없는 환경에서도 리스크 엔진의 모든 기능을 테스트할 수 있습니다.
*   **재현성 확보**: 동일한 시나리오 YAML 파일만 있다면 누구나 동일한 부하 상황에서의 리스크 리포트를 얻을 수 있습니다.
*   **극단적 상황 테스트**: 실제 운영 중에는 재현하기 힘든 '초고부하(Flash Sale)', '시스템 노후화' 등의 상황을 자유롭게 설정할 수 있습니다.

---

## 2. 시나리오 정의 (YAML Scenario)

샌드박스 생성을 위해서는 실험의 설계도 역할을 하는 YAML 파일이 필요합니다.

### 상세 필드 설명 (`scenario.yaml`)

| 필드 그룹 | 필드명 | 설명 |
| :--- | :--- | :--- |
| **기본 정보** | `experiment_name` | 생성될 SQLite 파일명 (예: `test` 입력 시 `test.db` 생성) |
| | `description` | 실험에 대한 상세 설명 |
| | `target_table` | 리스크 분석의 주 대상이 될 테이블명 |
| **Mock PG 상태** | `pg_state` | 분석 시점의 현재 DB 상태 (TPS, 커넥션, 테이블 크기 등) |
| **히스토리 프로필** | `days` | 생성할 과거 데이터 기간 (보통 7일 권장) |
| | `base_tps` / `peak_tps` | 평시 및 피크 시간대의 기준 TPS |
| | `weekly_pattern` | 주말에 트래픽이 40% 감소하는 패턴 적용 여부 |
| | `noise_variance` | 데이터의 불규칙성(Noise) 강도 (0.0 ~ 1.0) |
| **특수 이벤트** | `events` | 특정 날짜/시간에 발생하는 돌발 트래픽 (Flash Sale 등) 정의 |

### 시나리오 예시
```yaml
experiment_name: "fintech_spike_research"
description: "Sudden flash sale spike scenario on Day 3"
target_table: "account_balances"

pg_state:
  table_size_mb: 5120
  active_connections: 45
  p99_time_ms: 15.5
  current_tps: 1200
  replication_lag_s: 0.2

sqlite_history:
  days: 7
  interval_minutes: 60
  base_tps: 800
  peak_tps: 2500
  weekly_pattern: true
  noise_variance: 0.1
  events:
    - name: "Flash Sale"
      start_day: 3
      start_hour: 14
      duration_h: 4
      multiplier: 3.5
```

---

## 3. 명령어 사용법 (Usage)

### 3.1. 샌드박스 생성 (`simulate`)
시나리오 파일을 읽어 SQLite 샌드박스 데이터베이스를 생성합니다.

```bash
migraguard simulate --scenario scenario_test.yaml --verbose
```
*   **출력**: `[OK] Simulation Sandbox created: fintech_spike_research.db`

### 3.2. 오프라인 분석 (`analyze --sandbox`)
생성된 샌드박스 파일을 실제 DB인 것처럼 속여서 리스크 분석을 수행합니다.

```bash
migraguard analyze migrations/001_safe_add_column_orders.sql --sandbox fintech_spike_research.db
```
*   이 명령은 실제 PostgreSQL에 연결을 시도하지 않습니다.
*   샌드박스에 주입된 과거 피크 데이터와 현재 Mock 상태를 기반으로 리포트를 생성합니다.

---

## 4. 수학적 모델링 상세 (Technical Details)

샌드박스 엔진은 보다 현실적인 데이터를 위해 다음과 같은 알고리즘을 사용합니다.

1.  **일간/주간 사인파 (Sine Wave)**:
    *   오전 9시부터 트래픽이 상승하여 오후 2시와 8시에 정점을 찍는 현실적인 곡선을 그립니다.
    *   `weekly_pattern` 활성 시 토/일요일 트래픽이 자동으로 하향 조정됩니다.
2.  **지표 상관관계 (Correlation)**:
    *   **P99 지연 시간**: TPS가 상승하면 지연 시간은 지수 함수적으로 증가합니다 ($P99 = \text{Base} \times e^{(\frac{tps}{peak}-1)}$).
    *   **활성 커넥션**: TPS에 비례하여 데이터베이스 세션 수가 선형적으로 증가하도록 모사합니다.
3.  **워크로드 주입**:
    *   단순 TPS 수치뿐만 아니라, `workload_snapshots` 테이블에 실제 쿼리 로그(SELECT, UPDATE, INSERT)를 배분하여 주입함으로써 쿼리별 영향도 분석 기능을 지원합니다.

---

## 5. 활용 시나리오

1.  **CI/CD 파이프라인 테스트**: 실제 스테이징 DB 없이도 마이그레이션 위험도를 체크하는 로직을 파이프라인에 통합할 수 있습니다.
2.  **연구 및 논문 데이터 확보**: 특정 변수(TPS, Lock Level) 변화에 따른 리스크 점수 추이를 정량적으로 분석할 수 있습니다.
3.  **데모 및 프레젠테이션**: DB 서버 구축 없이 가벼운 SQLite 파일만으로 MigraGuard의 모든 분석 과정을 시연할 수 있습니다.
