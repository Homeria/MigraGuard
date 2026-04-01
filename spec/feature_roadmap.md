# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap)

## 1. 기능 분할 전략 (Feature Breakdown)
전체 시스템은 **정적 분석(Static)**, **동적 수집(Dynamic)**, **평가 엔진(Engine)**, **리포팅(Reporter)**의 4가지 핵심 레이어로 나뉩니다. 각 레이어는 상호 의존성을 최소화하여 독립적인 개발 및 테스트가 가능하도록 분할합니다.

### [Phase 1] Core Parser (정적 분석)
- **feat/parser**: SQL AST 분석을 통한 테이블 명, 컬럼 명, 락 레벨(AccessExclusive 등) 추출 기능.
- **Goal**: 입력된 SQL이 DB에 어떤 영향을 주는지 "이론적"으로 파악.
- **Checkpoints**:
  - [ ] **1.1 환경 설정**: `pg_query_go` 의존성 추가 및 `internal/parser/ast.go` 기본 구조 설계.
  - [ ] **1.2 AST 파싱**: SQL 문자열을 AST(JSON/Struct)로 변환하는 기초 함수 구현.
  - [ ] **1.3 타겟 추출**: `ALTER TABLE`, `CREATE INDEX`, `DROP TABLE` 등 주요 DDL에서 대상 테이블 명 추출.
  - [ ] **1.4 락 매핑**: 각 DDL 액션(Add Column, Rename, Drop)에 따른 PostgreSQL Lock Level 매핑 로직 구현.
  - [ ] **1.5 결과 정규화**: 분석 결과를 `AnalysisResult` 구조체로 반환하고 단위 테스트(Unit Test) 완료.

### [Phase 2] Data Foundation (데이터 기반)
- **feat/db-adapter**: PostgreSQL(운영) 및 SQLite(로컬 시계열) 연결 및 스키마 관리.
- **feat/collector**: `pg_stat_statements` 스냅샷을 주기적으로 SQLite에 저장하는 백그라운드 워커.
- **Goal**: 위험도 산출을 위한 "실제 운영 데이터" 확보.

### [Phase 3] Risk Engine (판단 로직)
- **feat/risk-engine**: 수집된 데이터와 파싱 결과를 결합하여 Risk Score 산출 및 Safe Window(최적 배포 시간) 계산 알고리즘.
- **Goal**: 데이터 기반의 정량적 배포 승인/거부 판정.

### [Phase 4] Reporter & Gatekeeper (출력 및 제어)
- **feat/reporter**: 터미널 UI(표, 그래프) 및 GitHub PR용 Markdown 리포트 생성.
- **feat/gatekeeper**: 위험도 임계치에 따른 Exit Code 제어 및 CI/CD 연동 최적화.

---

## 2. 브랜치 관리 및 구현 순서 (Branching & Roadmap)

브랜치는 `develop`을 중심으로 각 `feat/` 브랜치를 생성하며, **데이터의 흐름(Input -> Storage -> Logic -> Output)** 순서에 따라 구현하는 것이 의존성 관리에 유리합니다.

### 권장 구현 순서 (Roadmap)

1.  **`feat/parser` (Phase 1)**
    - 가장 먼저 선행되어야 함. 어떤 테이블을 분석할지 알아야 데이터 수집 대상을 필터링할 수 있음.
2.  **`feat/db-adapter` & `feat/collector` (Phase 2)**
    - 엔진이 돌아가기 위한 원천 데이터(Source of Truth)를 구축.
3.  **`feat/risk-engine` (Phase 3)**
    - 파서와 컬렉터가 완성된 시점에서 두 데이터를 결합하는 핵심 로직 구현.
4.  **`feat/reporter` (Phase 4)**
    - 최종 산출물을 사용자에게 보여주는 단계.
5.  **`feat/cli-core` (최종 통합)**
    - 각 모듈을 `cmd/migraguard` 명령어에 연결하고 전체적인 UX(설정 파일 처리 등)를 마감.

### 브랜치 흐름 예시
```text
main
  └── develop
        ├── feat/parser (Merged to develop)
        ├── feat/db-adapter (Merged to develop)
        ├── feat/collector (Merged to develop)
        ├── feat/risk-engine (Merged to develop)
        └── feat/reporter (Merged to develop)
```

---

## 3. 향후 확장성 고려사항 (Next Step)
- **feat/multi-db**: MySQL 등 타 RDBMS 지원 확장.
- **feat/gui-dashboard**: SQLite 데이터를 시각화하는 간단한 웹 대시보드.
