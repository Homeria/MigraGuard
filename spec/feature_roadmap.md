# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap)

## 1. 기능 분할 전략 (Feature Breakdown)

### [Phase 1] Core Parser (정적 분석) - ✅ 완료
- **feat/parser**: SQL AST 분석을 통한 테이블 명, 컬럼 명, 락 레벨 추출 기능.
- **Checkpoints**:
  - [x] **1.1 환경 설정**: `pg_query_go` 의존성 추가 및 `internal/parser/ast.go` 기본 구조 설계.
  - [x] **1.2 AST 파싱**: SQL 문자열을 AST로 변환하는 기초 함수 구현 및 `gcc` 환경 검증.
  - [x] **1.3 타겟 추출**: `ALTER`, `CREATE`, `INDEX` 등 주요 DDL에서 대상 테이블 명 추출.
  - [x] **1.4 락 매핑**: `CONCURRENTLY` 유무 등에 따른 PostgreSQL Lock Level 매핑 로직 구현.
  - [x] **1.5 결과 정규화**: 컬럼 추출 추가 및 전체 케이스 단위 테스트(Unit Test) 완료.

### [Phase 2] Data Foundation (데이터 기반) - ✅ 완료
- **feat/db-adapter & feat/collector**: 운영 DB 연결 및 로컬 시계열 데이터 저장소 구축.
- **Checkpoints**:
  - [x] **2.1 DB 연결 (PostgreSQL)**: `jackc/pgx`를 사용하여 운영 DB 연결 및 `pg_stat_statements` 조회 기능.
  - [x] **2.2 로컬 저장소 (SQLite)**: SQLite 스키마 설계 및 WAL 모드 기반의 시계열 데이터 저장소 구축.
  - [x] **2.3 백그라운드 컬렉터**: 고루틴 Ticker를 이용한 주기적인 워크로드 스냅샷 캡처 로직.
  - [x] **2.4 데이터 액세스 레이어**: 특정 테이블의 과거 트래픽 통계를 조회하는 Repository 구현.

### [Phase 3] Risk Engine (판단 로직) - 🛠️ 다음 작업
- **feat/risk-engine**: 데이터 결합 및 위험도 산출 알고리즘 구현.
- **Checkpoints**:
  - [ ] **3.1 리스크 알고리즘**: `Traffic * Lock * Wait` 기반의 정량적 Risk Score 산출 로직.
  - [ ] **3.2 통계 분석 엔진**: 중앙값(Median) 기반의 트래픽 밀도 분석 및 Safe Window 추천 기능.
  - [ ] **3.3 실시간 모니터링 연동**: `pg_stat_activity`를 통한 현재 실행 중인 장기 쿼리 영향도 평가.
  - [ ] **3.4 엔진 통합**: Parser 결과와 DB 데이터를 결합하는 핵심 엔진 API 완성.

### [Phase 4] Reporter & Gatekeeper (출력 및 제어)
- **feat/reporter & feat/gatekeeper**: 결과 시각화 및 배포 프로세스 제어.
- **Checkpoints**:
  - [ ] **4.1 콘솔 리포터**: 터미널 내 표(Table) 및 간단한 그래프(ASCII Chart) 출력 기능.
  - [ ] **4.2 마크다운 생성기**: GitHub PR 코멘트용 분석 리포트 생성 로직.
  - [ ] **4.3 게이트키퍼**: 위험도 임계치 초과 시 Exit Code 제어 및 CI/CD 연동 인터페이스.

---

## 2. 브랜치 관리 및 구현 순서 (Branching & Roadmap)

브랜치는 `develop`을 중심으로 각 `feat/` 브랜치를 생성하며, **데이터의 흐름(Input -> Storage -> Logic -> Output)** 순서에 따라 구현합니다.

### 브랜치 흐름 예시
```text
main
  └── develop
        ├── feat/parser (Merged ✅)
        ├── feat/db-adapter (Merged ✅)
        ├── feat/collector (Merged ✅)
        ├── feat/risk-engine (Next 🛠️)
        └── feat/reporter
```

---

## 3. 향후 확장성 고려사항 (Next Step)
- **feat/multi-db**: MySQL 등 타 RDBMS 지원 확장.
- **feat/gui-dashboard**: SQLite 데이터를 시각화하는 간단한 웹 대시보드.
