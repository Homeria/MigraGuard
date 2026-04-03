# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap)

## 1. 기능 분할 전략 (Feature Breakdown)

### [Phase 1] Core Parser (정적 분석) - ✅ 완료
- **feat/parser**: SQL AST 분석을 통한 테이블 명, 컬럼 명, 락 레벨 및 $F_{rewrite}$ 추출.
- **Checkpoints**:
  - [x] **1.1 환경 설정**: `pg_query_go` 의존성 추가 및 `internal/parser/ast.go` 기본 구조 설계.
  - [x] **1.2 AST 파싱**: SQL 문자열을 AST로 변환하는 기초 함수 구현 및 `gcc` 환경 검증.
  - [x] **1.3 타겟 추출**: `ALTER`, `CREATE`, `INDEX` 등 주요 DDL에서 대상 테이블 명 추출.
  - [x] **1.4 락 매핑**: `CONCURRENTLY` 유무 등에 따른 PostgreSQL Lock Level 매핑 로직 구현.
  - [x] **1.5 재기록 감지**: **v3.0 고도화** - `ALTER COLUMN TYPE` 등 테이블 재기록($F_{rewrite}$) 유발 구문 판별 로직 추가.

### [Phase 2] Data Foundation (데이터 기반) - ✅ 완료
- **feat/db-adapter & feat/collector**: 운영 DB 연결 및 v3.0 동적 지표 수집 체계 구축.
- **Checkpoints**:
  - [x] **2.1 DB 연결 (PostgreSQL)**: `jackc/pgx`를 사용하여 운영 DB 연결 및 `pg_stat_statements` 조회 기능.
  - [x] **2.2 로컬 저장소 (SQLite)**: SQLite `table_metrics` 테이블 확장 및 시계열 데이터 저장소 구축.
  - [x] **2.3 백그라운드 컬렉터**: **v3.0 고도화** - 주기적인 워크로드 스냅샷 및 동적 지표($S_{table}$, $Lag_{repl}$, $C_{active}$, $\lambda$) 캡처 루틴 통합.
  - [x] **2.4 데이터 액세스 레이어**: 특정 테이블의 실시간/과거 통계를 조회하는 Repository 구현.

### [Phase 3] Risk Engine (판단 로직) - ✅ 완료
- **feat/risk-engine**: 대기행렬 이론 기반의 v3.0 수학적 모델 구현.
- **Checkpoints**:
  - [x] **3.1 리스크 알고리즘**: **v3.0 고도화** - `T_ddl`, `T_block`, `C_peak` 기반의 정량적 Risk Score 산출 로직 구현.
  - [x] **3.2 회복 시간 분석**: 큐 스파이크 해소 시간($T_{rec}$) 및 영구 장애 예측 기능 추가.
  - [x] **3.3 실시간 모니터링 연동**: `pg_stat_activity`를 통한 현재 실행 중인 장기 쿼리 영향도 평가 통합.
  - [x] **3.4 엔진 통합**: Parser 결과와 DB 동적 지표를 결합하는 핵심 엔진 API 완성.

### [Phase 4] Reporter & Gatekeeper (출력 및 제어) - ✅ 완료
- **feat/reporter & feat/gatekeeper**: 결과 시각화 및 배포 프로세스 제어.
- **Checkpoints**:
  - [x] **4.1 콘솔 리포터**: **v3.0 고도화** - $RiskScore$, $T_{block}$, $C_{peak}$, $T_{rec}$ 등 정량적 지표의 터미널 리포팅 구현.
  - [x] **4.2 게이트키퍼**: 위험도 임계치 초과(`Danger`) 시 Exit Code 1 제어 및 CI/CD 연동 인터페이스 완성.
  - [ ] **4.3 마크다운 생성기**: (Next Step) GitHub PR 코멘트용 분석 리포트 생성 로직.

---

## 2. 향후 확장성 고려사항 (Next Step)
- **feat/config-manager**: $Disk_{IO}$, $\mu_{max}$ 등 엔진 상수를 외부 설정 파일(`migraguard.yaml`)로 분리.
- **feat/multi-db**: MySQL 등 타 RDBMS 지원 확장.
- **feat/gui-dashboard**: SQLite 데이터를 시각화하는 간단한 웹 대시보드.
