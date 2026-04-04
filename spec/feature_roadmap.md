# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap v3.1)

## 1. 기능 분할 전략 (Feature Breakdown)

### [Phase 1] Core Parser (정적 분석) - ✅ 완료
- **feat/parser**: SQL AST 분석을 통한 테이블 명, 컬럼 명, 락 레벨 및 $F_{rewrite}$ 추출.
- (기존 완료 항목 유지)

### [Phase 2] Data Foundation (v3.1 에이전트 전환) - ✅ 완료
- **feat/agent**: 상시 수집 에이전트(Agent) 독립화 및 시계열 데이터 관리.
- **Checkpoints**:
  - [x] **2.1 DB 연결 & SQLite 저장소**: 기초 연결 로직 완료.
  - [x] **2.2 에이전트 독립 실행**: `migraguard agent` 커맨드 및 백그라운드 무한 루프 구현.
  - [x] **2.3 데이터 보존 정책 (Retention)**: 7일 경과 데이터 자동 삭제 및 `VACUUM` 로직 완료.
  - [x] **2.4 헬스체크**: 에이전트 수집 상태 모니터링 및 로깅 강화.

### [Phase 3] Risk Engine (v3.1 Baseline 모델) - ✅ 완료
- **feat/risk-engine-v3.1**: 3초 대기 제거 및 시계열 데이터 기반 즉각 분석.
- **Checkpoints**:
  - [x] **3.1 기본 큐잉 모델**: 수학적 수식 구현 완료.
  - [x] **3.2 시계열 지표 쿼리**: SQLite `avg()`, `max()` 기반 베이스라인($\lambda$) 추출 로직 완료.
  - [x] **3.3 Safe Window 추천**: 24시간 트래픽 패턴 분석을 통한 최적 배포 시간대 산출.
  - [x] **3.4 실시간-과거 가중치 결합**: $\lambda_{final}$ 산출 로직 고도화 완료.

### [Phase 4] Reporter & Gatekeeper (v3.1 통합) - ✅ 완료
- **feat/reporter-v3.1**: 분석 결과의 컨텍스트 강화 및 CI/CD 최적화.
- **Checkpoints**:
  - [x] **4.1 콘솔 리포터**: 기본 출력 로직 완료.
  - [x] **4.2 컨텍스트 리포팅**: "현재 트래픽이 평소 대비 20% 높습니다" 등의 비교 문구 추가.
  - [x] **4.3 마크다운 리포터**: GitHub PR용 트래픽 추이 및 요약 테이블을 포함한 리포트 생성 완료.

### [Phase 5] Configuration & Infrastructure - ✅ 완료
- **feat/configuration**: 상수 외부화 및 설정 파일(`migraguard.yaml`) 도입.
- **infra/docker-setup**: 멀티 스테이지 빌드 및 공유 볼륨을 통한 에이전트-CLI 연동 검증.
- **Checkpoints**:
  - [x] **5.1 Viper 통합**: YAML 기반 환경 변수 및 설정 파일 연동.
  - [x] **5.2 Dockerization**: 에이전트와 분석기가 데이터를 공유하는 표준 도커 환경 구축.

---

## 2. 향후 릴리즈 계획 (Post v3.1)

### [Phase 6] Integrity & Stability (준비 중)
- **feat/schema-validation**: `information_schema`를 통한 테이블/컬럼 존재 여부 사전 체크.
- **feat/debug-mode**: `--verbose` 플래그를 통한 분석 중간 과정 상세 노출.
- **feat/unit-testing**: 리스크 엔진 및 파서에 대한 시나리오 기반 유닛 테스트 강화.
