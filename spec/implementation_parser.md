# MigraGuard Phase 1: Core Parser 구현 리포트 (v3.0 고도화)

## 1. 구현 개요
- **목적**: 입력된 SQL(DDL)을 분석하여 영향을 받는 테이블, 컬럼, 락 레벨 및 **테이블 재기록($F_{rewrite}$)** 여부를 식별.
- **핵심 라이브러리**: `github.com/pganalyze/pg_query_go/v5`.

## 2. 고도화된 데이터 구조 (`internal/parser/ast.go`)
### `AnalysisResult` 구조체 확장
```go
type AnalysisResult struct {
    TableName       string
    LockLevel       LockLevel
    Operation       string   // CREATE, ALTER, INDEX 등
    Columns         []string
    RewriteRequired bool     // F_rewrite: 테이블 풀 스캔 및 복사 발생 여부 판별
}
```

## 3. 핵심 로직 상세 (v3.0 핵심)
- **$F_{rewrite}$ 판단 로직**:
    - `ALTER TABLE`의 서브 명령(`Subtype`)을 정밀 분석.
    - `AT_AlterColumnType`, `AT_SetNotNull`, 기본값이 있는 `AT_AddColumn` 등을 재기록 발생 작업으로 식별.
- **락 매핑 정교화**: `CONCURRENTLY` 유무에 따른 `ShareUpdateExclusiveLock` 산출 로직 유지.

## 4. 향후 과제
- PostgreSQL 버전에 따른 재기록 발생 여부의 미세한 차이(예: v12+ `SET NOT NULL` 최적화) 대응 고도화.
