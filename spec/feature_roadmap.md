# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap v3.1)

## 1. 기능 분할 전략 (Feature Breakdown)

### [Phase 1] Core Parser (정적 분석) - ✅ 완료
- **feat/parser**: SQL AST 분석을 통한 테이블 명, 컬럼 명, 락 레벨 및 $F_{rewrite}$ 추출.
- (기존 완료 항목 유지)

### [Phase 2] Data Foundation (v3.1 에이전트 전환) - 🚧 진행 중
- **feat/agent**: 상시 수집 에이전트(Agent) 독립화 및 시계열 데이터 관리.
- **Checkpoints**:
  - [x] **2.1 DB 연결 & SQLite 저장소**: 기초 연결 로직 완료.
  - [ ] **2.2 에이전트 독립 실행**: `migraguard agent` 커맨드 구현 및 백그라운드 무한 루프.
  - [ ] **2.3 데이터 보존 정책 (Retention)**: 7일 경과 데이터 자동 삭제 및 `VACUUM` 로직.
  - [ ] **2.4 헬스체크**: 에이전트 수집 상태 모니터링 및 로깅 강화.

### [Phase 3] Risk Engine (v3.1 Baseline 모델) - 🚧 진행 중
- **feat/risk-engine-v3.1**: 3초 대기 제거 및 시계열 데이터 기반 즉각 분석.
- **Checkpoints**:
  - [x] **3.1 기본 큐잉 모델**: 수학적 수식 구현 완료.
  - [ ] **3.2 시계열 지표 쿼리**: SQLite `avg()`, `max()` 기반 베이스라인($\lambda$) 추출 로직.
  - [ ] **3.3 Safe Window 추천**: 24시간 트래픽 패턴 분석을 통한 최적 배포 시간대 산출.
  - [ ] **3.4 실시간-과거 가중치 결합**: $\lambda_{final}$ 산출 로직 고도화.

### [Phase 4] Reporter & Gatekeeper (v3.1 통합) - 🚧 예정
- **feat/reporter-v3.1**: 분석 결과의 컨텍스트 강화 및 CI/CD 최적화.
- **Checkpoints**:
  - [x] **4.1 콘솔 리포터**: 기본 출력 로직 완료.
  - [ ] **4.2 컨텍스트 리포팅**: "현재 트래픽이 평소 대비 20% 높습니다" 등의 비교 문구 추가.
  - [ ] **4.3 마크다운 리포터**: GitHub PR용 트래픽 추이 차트 포함 리포트 생성.

---

## 2. 브랜치 전략 추천 (Gitflow 기반)

현재 프로젝트는 대규모 아키텍처 전환(v3.0 -> v3.1) 단계에 있으므로, 안정적인 관리를 위해 다음과 같은 브랜치 구조를 추천합니다.

### 2.1. 주요 브랜치
- **`main`**: 상용 수준의 안정적인 코드만 병합. (v3.0 상태 유지)
- **`develop`**: 차기 버전(v3.1) 개발을 위한 통합 브랜치.
- **`release/v3.1`**: v3.1 정식 배포 전 최종 버그 수정 및 안정화 브랜치.

### 2.2. 기능 개발 브랜치 (Feature Branches) - `develop`에서 분기
1.  **`feat/agent-mode`**: `agent` 커맨드 및 상시 수집 로직 구현.
2.  **`feat/analyze-instant`**: `analyze` 명령의 `time.Sleep` 제거 및 SQLite 조회 로직 수정.
3.  **`feat/risk-v3.1-logic`**: 시계열 가중치 기반 리스크 산출 수식 적용.
4.  **`infra/docker-volume`**: Agent-CLI 간 데이터 공유를 위한 Docker 설정.

### 2.3. 작업 흐름 예시
1. `develop` 브랜치에서 `feat/agent-mode` 분기.
2. 에이전트 기능 완료 후 `develop`에 PR 및 병합.
3. `develop`의 최신 코드를 기반으로 `feat/analyze-instant` 작업 시작.
4. 모든 v3.1 기능이 `develop`에 모이면 `release/v3.1` 생성 및 최종 테스트.
5. `main`으로 병합하여 정식 v3.1 릴리즈.
