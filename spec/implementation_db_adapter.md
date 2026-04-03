# MigraGuard Phase 2: DB Adapter & Storage 구현 리포트 (v3.0 고도화)

## 1. PostgreSQL 동적 지표 수집 (`internal/db/workload.go`)
- **목적**: v3.0 위험도 모델에 필요한 정량적 실시간 지표 수집.
- **주요 메서드 (`GetTableDynamicMetrics`)**:
    - **$S_{table}$**: `pg_total_relation_size()`를 통한 물리 크기 조회.
    - **$Lag_{repl}$**: `pg_stat_replication`을 통한 WAL 동기화 지연 시간 조회.
    - **$C_{active}$**: `pg_stat_activity` 및 `LIKE` 패턴 매칭을 통한 활성 커넥션 조회.
    - **$T_{p99}$ / $\lambda$**: `pg_stat_statements` 및 `PERCENTILE_CONT`를 활용한 성능 통계 추출.

## 2. SQLite 시계열 저장소 확장 (`internal/db/activity.go`)
- **신규 테이블**: `table_metrics` (v3.0 동적 지표 전용 시계열 저장소).
- **데이터 영속화**: `SaveTableMetrics`를 통해 수집된 지표를 SQLite에 적재.

## 3. 백그라운드 컬렉터 통합 (`internal/db/collector.go`)
- **동적 대상 관리**: `AddTargetTable`을 통해 모니터링 대상 테이블 관리.
- **수집 파이프라인**: 워크로드 스냅샷 수집 루틴에 테이블별 동적 지표 수집 로직 통합.

## 4. 향후 과제
- 수집 주기(Interval) 및 유지 관리 정책(Retention)을 설정 파일에서 관리.
