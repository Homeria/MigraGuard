# MigraGuard v3.1: 상시 수집 에이전트(Agent) 구현 리포트

본 문서는 `feat/agent-mode` 브랜치에서 구현된 상시 수집 에이전트의 상세 내역을 기록합니다.

## 1. 구현 개요
- **목적**: 운영 DB 옆에서 24/7 가동하며 실시간 워크로드 지표를 SQLite 시계열 저장소에 누적하고 관리함.
- **핵심 모듈**: `cmd/migraguard/agent.go`, `internal/db/collector.go`, `internal/db/activity.go`

## 2. 주요 로직 (Logic ID)

### [L-A01] 에이전트 독립 실행 구조
- `cmd/migraguard/agent.go`를 통해 독립적인 서브커맨드로 구현.
- `cobra.Command`를 사용하여 `--db`, `--sqlite`, `--interval`, `--retention` 옵션 지원.
- OS Signal(`SIGTERM`, `os.Interrupt`)을 감지하여 Graceful Shutdown 처리.

### [L-A02] 주기적 지표 수집 및 저장
- `internal/db/collector.go`의 `collect()` 메서드에서 다음을 수행:
  - `pg_stat_statements` 스냅샷 캡처 및 SQLite 저장.
  - 대상 테이블별 동적 지표(Size, Lag, Connections) 수집 및 저장.

### [L-A03] 데이터 보존 정책 (Retention Policy)
- `internal/db/activity.go`에 `PurgeOldSnapshots(retentionDays)` 메서드 구현.
- 설정된 일수(기본 7일)가 지난 데이터를 `DELETE` 쿼리로 삭제.
- `VACUUM` 명령을 통해 SQLite 파일의 물리적 공간을 최적화하여 디스크 비대화 방지.

### [L-A04] 에이전트 헬스체크 및 로깅
- 수집 성공 시 스냅샷 개수와 대상 테이블명을 로그로 출력.
- DB 연결 실패 시 에러 로그를 남기고 다음 주기에 재시도하도록 설계.

## 3. 실행 예시
```bash
# 기본 실행 (60초 간격, 7일 보존)
migraguard agent --db "postgres://user:pass@localhost:5432/dbname"

# 커스텀 설정 실행
migraguard agent --db "..." --interval 10 --retention 30 --sqlite "./data/metrics.db"
```

## 4. 향후 과제
- 분석 CLI(`analyze`)에서 에이전트가 쌓은 데이터를 즉시 활용하도록 로직 통합.
- 에이전트 실행 상태를 모니터링하기 위한 간단한 헬스체크 API 추가 검토.
