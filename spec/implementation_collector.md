# MigraGuard Phase 2: Background Collector 구현 리포트

본 문서는 `feat/collector` 단계에서 진행된 백그라운드 데이터 수집 엔진의 구현 상세 내역을 기록합니다.

## 1. 구현 개요
- **목적**: 메인 프로세스의 블로킹 없이 주기적으로 운영 DB 데이터를 로컬 저장소로 전송.
- **핵심 모듈**: `internal/db/collector.go`

## 2. 수집기 구조 (`Collector` Struct)
```go
type Collector struct {
    pg       *PostgresAdapter // 데이터 소스
    sqlite   *SQLiteAdapter   // 저장 대상
    interval time.Duration    // 수집 주기
    stopChan chan struct{}    // 종료 신호 채널
}
```

## 3. 핵심 동작 로직
- **생명주기 관리**:
    - `Start(ctx)`: 별도의 고루틴에서 `time.Ticker`를 사용하여 주기적 루프 실행.
    - `Stop()`: 채널 닫기를 통한 안전한 고루틴 종료(Graceful Shutdown).
- **수집 파이프라인 (`collect`)**:
    1. `PostgresAdapter.FetchWorkload`를 호출하여 현재 쿼리 통계 스냅샷 획득.
    2. 데이터가 존재할 경우 `SQLiteAdapter.SaveSnapshots`를 통해 일괄 저장(Bulk Insert).
    3. 트랜잭션을 사용하여 SQLite 쓰기 무결성 보장.

## 4. 동시성 제어 및 에러 처리
- **Context 지원**: `ctx.Done()` 감지를 통해 부모 프로세스 종료 시 즉각적인 자원 해제.
- **로깅**: 수집 성공/실패 여부를 표준 로그로 출력하여 모니터링 가능하도록 구현.

## 5. 향후 과제
- 수집 주기(Interval)를 설정 파일(`config`)에서 동적으로 로드하도록 개선.
- 수집 실패 시 재시도(Retry) 전략 수립.
