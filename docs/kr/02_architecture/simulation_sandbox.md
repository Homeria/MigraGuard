# 🧪 샌드박스 파일 기반 시뮬레이션 아키텍처

이 문서는 학술 연구 및 통계적 검증에 최적화된 MigraGuard v3.9의 **시뮬레이션 샌드박스** 설계 및 구현 상세를 설명합니다.

---

## 1. 개요

시뮬레이션 샌드박스는 연구자가 실제 PostgreSQL 인스턴스 없이도 통제된 환경에서 마이그레이션 리스크를 평가할 수 있게 합니다. YAML로 시나리오를 정의하면, MigraGuard는 독립적인 SQLite 샌드박스 파일을 생성하고 가상화된 데이터베이스 상태를 기반으로 리스크 분석을 수행합니다.

v3.9 릴리스에서는 수학적 데이터 생성기(Profiler)를 완전히 독립시키고, 동시성 경합이 없는 로컬 난수 소스(`rng`) 및 `tx.Prepare` 오류 방지 안전 장치들을 탑재하여 모의 생성의 신뢰성을 극대화했습니다.

### 핵심 목표
- **재현성 (Reproducibility)**: 동일한 시나리오 YAML과 SQL에 대해 항상 일정한 결과 보장.
- **격리성 (Isolation)**: 운영용 메트릭이 담긴 `migraguard.db`와의 간섭 방지.
- **기동성 (Portability)**: 표준 SQLite 도구로 감사할 수 있는 물리적 `.db` 파일 생성.
- **통계 분석 및 시각화**: 24h 개별 파동 시각화(Heatmap) 및 종합 마스터 리포트 PNG 차트 출력을 위한 인프라 제공.

---

## 2. v3.9 선언적 리얼리스틱 시나리오 (YAML)

v3.9 시나리오는 비대칭 왜곡률, 피크 시프트 offset, 주말 트래픽 감소 비율, 일 단위 볼륨 스케일링 요소 등을 추가로 제어할 수 있는 고수준의 물리 속성을 내장하고 있습니다.

```yaml
# 예시: experiments/scenarios/04_sine_daily_cycle.yaml
experiment_name: "sine_daily_cycle"
description: "Realistic daily traffic swing simulation with high skewness"
target_table: "orders"

# 가상 PostgreSQL 메트릭 상태 (Now)
pg_state:
  table_size_mb: 2048
  replication_lag_s: 0
  active_connections: 50
  p99_time_ms: 10.0
  current_tps: 150.0

# v3.9 고도화된 SQLite 통계적 이력 생성 규칙
sqlite_history:
  days: 7
  interval_minutes: 30
  base_tps: 50.0
  peak_tps: 500.0
  noise_variance: 0.10
  asymmetric_skew: 0.25      # 비대칭 시간 왜곡률
  peak_shift_hours: 1.5      # 피크 시간 편차 오프셋
  weekly_pattern: true       # 주말 트래픽 40% 감소 패턴
  events:                    # 특정 시각 트래픽 이벤트 돌발 변동
    - name: "evening-burst"
      start_day: 3
      start_hour: 18
      duration_h: 4
      multiplier: 1.8
```

---

## 3. 구현 단계 및 아키텍처 결합 격리

### 1단계: 통계 모델(Profiler) 격리
- `DefaultWorkloadProfiler` 구조체를 `sandbox_profiler.go` 로 격리하여 순수 수학적 파동 계산(Sine, Time-warped Skew, Noise)만을 전담 처리하도록 독립시켰습니다.
- 글로벌 패키지 난수 경합을 거세하고, 동시성 안정성을 보장하기 위해 `rand.New(rand.NewSource(now.UnixNano()))` 를 활용한 스레드-세이프 로컬 난수 발생기를 도입했습니다.

### 2단계: 가상 PG 어댑터 (`VirtualPGAdapter`)
- `PostgresClient` 인터페이스를 구현하는 `VirtualPGAdapter` (`sqlite_virtual.go`) 생성.
- YAML의 `pg_state` 값을 어댑터 응답(테이블 크기, 커넥션, P99, TPS 등)으로 완벽 매핑.
- `Scan` 실패 및 `rows.Next()` 루프 종료 시 `rows.Err()`을 의무 확인하여 데이터 유실 없는 무손실 가상화를 실현했습니다.

### 3단계: 연구 지표 로거 & L2/L3 오케스트레이터 연동
- L1 마이크로 스크립트 실행으로 도출되는 CSV 원시 데이터를 종합 수집해 single execution 타임스탬프인 `experiments/reports/batch_runs/[TIMESTAMP]/` 하위에 일괄 패키징하고, 24h 히트맵과 마스터 PNG 차트를 자동 생성하는 파이프라인 연계를 구현했습니다.

---

## 4. CLI 사용법 및 워크플로우

### 1) 단일 샌드박스 시뮬레이션 기동
시나리오 YAML 설정 파일을 바탕으로 로컬 가상 샌드박스 데이터베이스를 즉시 시딩 구축합니다:
```bash
./build/migraguard simulate --scenario experiments/scenarios/04_sine_daily_cycle.yaml --force
```

### 2) L2 시나리오 파이프라인 오케스트레이션
지정된 단일 시나리오를 구동한 후 가상 DDL 마이그레이션 정적/동적 리스크 종합 분석을 일괄 배치 처리하여 마스터 Heatmap 차트 보고서까지 일거에 생성합니다:
```bash
bash scripts/sh/run_scenario_pipeline.sh experiments/scenarios/04_sine_daily_cycle.yaml
```

### 3) L3 글로벌 벤치마크 파이프라인
확장된 20개 시나리오 풀 전체에 대해 dynamic 모의 샌드박스 배치 시뮬레이션을 전수 기동하여 다차원 리서치 분석 결과물을 단 한 줄의 명령어로 패키징 추출합니다:
```bash
bash scripts/sh/run_global_pipeline.sh
```
