# MigraGuard 기술 사양 및 아키텍처 (상세)

## 1. 시스템 동작 흐름 (Core Engine Flow)
1. **백그라운드 수집 (Background Collection):**
   - 고루틴 기반 데몬이 `robfig/cron` 또는 Ticker를 사용하여 주기적으로 운영 DB의 `pg_stat_statements` 스냅샷을 캡처합니다.
   - 캡처된 데이터는 `internal/db/workload.go`를 통해 로컬 SQLite 저장소에 누적됩니다.
2. **분석 단계 (Analyze Phase):**
   - 사용자가 `analyze` 명령으로 마이그레이션 SQL을 입력합니다.
   - `internal/parser/ast.go`가 SQL을 AST로 변환하여 타겟 테이블을 식별하고, 해당 테이블에 필요한 Lock Level을 판별합니다.
3. **위험도 평가 (Risk Evaluation):**
   - SQLite에 저장된 시계열 트래픽 데이터를 분석하여, 해당 테이블의 평소 트래픽 밀도를 파악합니다.
   - 중앙값(Median) 연산을 통해 '최적의 안전 배포 시간대(Safe Window)'를 추천합니다.
   - 현재 트래픽이 임계치를 초과하거나 락 레벨이 높은 경우 `Risk Score`를 산출합니다.
4. **결과 리포팅 및 게이트키핑:**
   - CLI 결과 출력 및 GitHub PR 봇 연동을 통한 마크다운 코멘트 작성.
   - 위험 수준이 'Danger'인 경우 프로세스를 중단시켜 배포를 차단합니다.

## 2. 주요 모듈 및 기술 상세
- **데이터 저장소 (Storage):** 외부 의존성을 줄이기 위해 로컬 SQLite(`mattn/go-sqlite3`)를 채택하여 WAL 모드로 시계열 데이터를 관리합니다.
- **파서 (Parser):** PostgreSQL 정식 C 파서를 포팅한 `pg_query_go`를 사용하여 최신 PostgreSQL 문법을 완벽히 지원합니다.
- **동시성 모델:** Go의 `channel`과 `goroutine`을 활용하여 메인 프로세스의 블로킹 없이 백그라운드 수집을 수행합니다.

## 3. 핵심 알고리즘: Risk Score 산출
- `Risk Score = (Traffic Intensity) * (Lock Weight) * (Wait Factor)`
  - `Traffic Intensity`: 대상 테이블을 참조하는 쿼리들의 TPS(Transaction Per Second).
  - `Lock Weight`: DDL 종류별 가중치 (예: `AccessExclusiveLock` = 1.0, `ShareUpdateExclusiveLock` = 0.3).
  - `Wait Factor`: 현재 진행 중인 장기 실행 쿼리(`pg_stat_activity`)의 영향도.
