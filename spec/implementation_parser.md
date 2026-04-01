# MigraGuard Phase 1: Core Parser 구현 리포트

본 문서는 `feat/parser` 브랜치에서 진행된 정적 분석 엔진의 구현 상세 내역을 기록합니다.

## 1. 구현 개요
- **목적**: 입력된 SQL(DDL)을 분석하여 영향을 받는 테이블, 컬럼 및 PostgreSQL 락 레벨을 식별.
- **핵심 라이브러리**: `github.com/pganalyze/pg_query_go/v5` (PostgreSQL 정식 파서 포팅 버전).

## 2. 주요 데이터 구조 (`internal/parser/ast.go`)
### `AnalysisResult` 구조체
```go
type AnalysisResult struct {
    TableName string   // 대상 테이블 명
    LockLevel LockLevel // 식별된 PostgreSQL 락 레벨
    Operation string    // 작업 유형 (CREATE, ALTER, DROP 등)
    Columns   []string  // 영향받는 컬럼 목록
}
```

### `LockLevel` 정의
PostgreSQL의 공식 락 계층 구조를 반영하여 정의되었습니다.
- `AccessExclusiveLock`: `ALTER`, `DROP`, `CREATE TABLE` 등 기본값.
- `ShareLock`: 일반 `CREATE INDEX`.
- `ShareUpdateExclusiveLock`: `CREATE INDEX CONCURRENTLY`.

## 3. 핵심 로직 상세
- **AST 순회 (`handleNode`)**: `pg_query.Node`를 탐색하여 각 구문 타입을 식별.
- **타겟 추출**: `stmt.Relation.Relname`을 통해 대상 테이블 명을 정확히 추출.
- **지능형 락 판별**: `CREATE INDEX`의 경우 `Concurrent` 플래그 유무에 따라 락 강도를 다르게 산출함.
- **컬럼 추출**: `ALTER TABLE`의 서브 명령(`Cmds`) 및 `CREATE TABLE`의 정의(`TableElts`)에서 컬럼명을 리스트업.

## 4. 검증 결과
- **테스트 파일**: `internal/parser/ast_test.go`
- **검증 케이스**:
  - `ALTER TABLE users ADD COLUMN age INT;` -> (users, ALTER, Access Exclusive, [age])
  - `CREATE INDEX CONCURRENTLY ...` -> (profile, CREATE INDEX, Share Update Exclusive, [])
  - 멀티 스테이트먼트 지원 확인.
- **환경 참고**: CGO 기반 라이브러리이므로 `gcc`가 설치된 환경에서만 `go test` 및 빌드가 가능함.

## 5. 향후 과제
- `DROP` 구문에서 복합 객체(여러 테이블 동시 삭제 등) 파싱 정교화.
- `ALTER TABLE` 내의 인덱스 생성 등 특수한 케이스 추가 매핑.
