# MigraGuard v3.1: 상시 수집 에이전트(Agent) 구현 명세서

본 문서는 `feat/agent` 단계에서 진행되는 상시 수집 에이전트의 구현 상세를 정의합니다.

## 1. 구현 개요
- **목적**: 운영 DB 옆에서 24/7 가동하며 실시간 워크로드 지표를 SQLite 시계열 저장소에 누적.
- **핵심 모듈**: `cmd/migraguard/agent.go`, `internal/db/collector.go`

## 2. 에이전트 구조 및 생명주기
- **실행 명령**: `migraguard agent --db "DSN" --interval 60s`
- **동작 루프**:
    - 별도의 고루틴이 아닌 메인 프로세스 루프로 동작 (독립 프로세스).
    - `os.Interrupt` 및 `SIGTERM` 감지 시 Graceful Shutdown 처리.
- **수집 주기**: 기본 60초(추천), 실시간성이 필요한 경우 1초 단위 설정 가능.

## 3. 핵심 기능 상세

### 3.1. 지표 수집 (Metrics Collection)
- **대상**: `pg_stat_statements` (TPS, 꼬리지연), `pg_stat_replication` (복제지연), `pg_stat_activity` (활성커넥션).
- **데이터 흐름**: PostgreSQL Fetch -> SQLite Insert (누적치 스냅샷).

### 3.2. 데이터 보존 정책 (Retention Policy)
- **자동 삭제**: 매 수집 사이클마다 또는 주기적으로 구형 데이터 삭제.
  ```sql
  DELETE FROM workload_snapshots WHERE created_at < datetime('now', '-7 days');
  ```
- **최적화**: 주기적인 `VACUUM` 명령 실행으로 SQLite 파일 크기 관리.

### 3.3. 안정성 및 복구
- **재연결 로직**: DB 연결 유실 시 지수 백오프(Exponential Backoff) 기반 재시도.
- **로그**: 수집 실패 원인 및 시스템 부하 상태를 정기적으로 로깅.

## 4. 향후 과제
- 에이전트의 CPU/Memory 점유율 최적화.
- 다중 DB 인스턴스 동시 수집 지원.
