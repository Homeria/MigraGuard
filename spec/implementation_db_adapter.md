# MigraGuard Phase 2: DB Adapter & Storage 구현 리포트

본 문서는 `feat/db-adapter` 단계에서 진행된 데이터베이스 연결 및 로컬 저장소 구축 상세 내역을 기록합니다.

## 1. PostgreSQL 어댑터 (`internal/db/connection.go`, `workload.go`)
- **목적**: 운영 DB 연결 관리 및 쿼리 워크로드 데이터 수집.
- **핵심 기술**: `github.com/jackc/pgx/v5/pgxpool`
- **주요 기능**:
    - **연결 풀링**: `pgxpool`을 사용하여 고성능 동시 연결 관리.
    - **워크로드 추출 (`FetchWorkload`)**: `pg_stat_statements` 뷰를 조회하여 쿼리별 호출 횟수, 실행 시간, I/O 통계 추출.
    - **테이블 통계 (`GetTableStats`)**: `pg_stat_user_tables`를 통해 특정 테이블의 실시간 스캔 횟수 파악.

## 2. SQLite 시계열 저장소 (`internal/db/activity.go`)
- **목적**: 수집된 워크로드 데이터를 로컬에 누적하여 시계열 분석 기반 마련.
- **핵심 기술**: `github.com/mattn/go-sqlite3` (WAL 모드 활성화)
- **스키마 설계 (`workload_snapshots`)**:
    - `timestamp`: 데이터 수집 시점 (시계열 분석용).
    - `query_id`, `query`: 쿼리 식별 및 텍스트.
    - `calls`, `total_time`: 누적 성능 지표.
- **최적화**: `timestamp` 및 `query_id`에 인덱스를 부여하여 대량 데이터 조회 성능 확보.

## 3. 데이터 분석 레이어 (Repository)
- **목적**: 저장된 스냅샷을 통계적으로 집계하여 리스크 판단 근거 제공.
- **주요 로직 (`GetHistoricalStats`)**:
    - 특정 테이블명이 포함된 모든 쿼리를 대상으로 평균 호출 횟수(`AVG`), 최대 호출 횟수(`MAX`), 총 실행 시간 집계.
    - 쿼리 패턴 매칭(`LIKE %tableName%`)을 통한 대상 식별.

## 4. 향후 과제
- `pg_stat_statements_reset()` 호출 주기와의 동기화 고려.
- SQLite 데이터 보존 기간(Retention Policy) 설정 로직 추가 필요.
