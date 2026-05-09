# 🧪 샌드박스 파일 기반 시뮬레이션 아키텍처

이 문서는 학술 연구 및 통계적 검증에 최적화된 MigraGuard v3.8의 **시뮬레이션 샌드박스** 설계 및 구현 상세를 설명합니다.

---

## 1. 개요

시뮬레이션 샌드박스는 연구자가 실제 PostgreSQL 인스턴스 없이도 통제된 환경에서 마이그레이션 리스크를 평가할 수 있게 합니다. YAML로 시나리오를 정의하면, MigraGuard는 독립적인 SQLite 샌드박스 파일을 생성하고 가상화된 데이터베이스 상태를 기반으로 리스크 분석을 수행합니다.

### 핵심 목표
- **재현성 (Reproducibility)**: 동일한 시나리오 YAML과 SQL에 대해 항상 일정한 결과 보장.
- **격리성 (Isolation)**: 운영용 메트릭이 담긴 `migraguard.db`와의 간섭 방지.
- **기동성 (Portability)**: 표준 SQLite 도구로 감사할 수 있는 물리적 `.db` 파일 생성.
- **통계 분석 (Statistical Analysis)**: 외부 처리(Python, Excel 등)를 위해 원시 리스크 지표를 CSV로 추출.

---

## 2. 선언적 시나리오 (YAML)

실험은 선언적 구문을 사용하여 정의됩니다. 이를 통해 실험 설정의 버전 관리가 가능해집니다.

```yaml
# 예시: experiments/spike_test.yaml
experiment_name: "High Traffic Spike"
description: "플래시 세일 중 트래픽 5배 증가 상황 시뮬레이션"
target_table: "orders"

# 가상 PostgreSQL 상태
pg_state:
  table_size_mb: 2048
  active_connections: 350
  p99_time_ms: 45.0
  current_tps: 3000.0

# SQLite 이력 생성 규칙
sqlite_history:
  pattern: "spike" # 패턴: steady, sine, spike, random
  days: 7
  base_tps: 500.0
  peak_tps: 4500.0
```

---

## 3. 구현 단계

### 1단계: 시뮬레이션 샌드박스 (`feat/simulation-sandbox`)
- 시나리오 정의를 위한 YAML 파서 구현.
- 각 실행마다 새로운 `{experiment_name}.db`를 생성하는 샌드박스 엔진 구축.
- SQLite 테이블을 채우기 위한 수학적 데이터 생성기(Sine, Spike 등) 구현.

### 2단계: 가상 PG 어댑터 (`feat/virtual-pg-adapter`)
- `PostgresClient` 인터페이스를 구현하는 `VirtualPGAdapter` 생성.
- YAML의 `pg_state` 값을 어댑터 응답(테이블 크기, 커넥션, P99)으로 매핑.
- `mg analyze --mock` 플래그를 통해 SDK가 가상 어댑터를 사용하도록 전환.

### 3단계: 연구 지표 로거 (`feat/research-csv-export`)
- 모든 5단계 리스크 지표($T_{ddl}, T_{block}, C_{peak}, T_{rec}$, 점수)를 캡처하는 구조화된 로거 구현.
- CLI에 CSV 내보내기 기능 추가.

---

## 4. 사용 워크플로우

1. **설계**: 시나리오 YAML 작성.
2. **실행**: `mg simulate --scenario path/to/yaml --sql path/to/sql` 실행.
3. **검증**: `sqlite3`를 사용하여 생성된 `.db` 파일 검토.
4. **분석**: 생성된 `results.csv`를 연구 논문이나 발표 자료로 활용.
