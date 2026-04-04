# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap v3.1)

## 1. 기능 분할 전략 (Feature Breakdown)

### [Phase 1] Core Parser (정적 분석) - ✅ 완료
- **feat/parser**: SQL AST 분석을 통한 테이블 명, 컬럼 명, 락 레벨 및 $F_{rewrite}$ 추출.

### [Phase 2~5] v3.1 Architecture & Infrastructure - ✅ 완료
- **feat/agent**: 상시 수집 에이전트(Agent) 독립화 및 시계열 데이터 관리.
- **feat/risk-engine-v3.1**: 3초 대기 제거 및 베이스라인 모델($\lambda_{final}$) 구축.
- **feat/reporter-v3.1**: 마크다운 리포터 및 GitHub PR 연동 최적화.
- **feat/configuration**: 상수 외부화 및 설정 파일(`migraguard.yaml`) 도입.
- **infra/docker-setup**: 멀티 스테이지 빌드 및 실전형 검증 환경 구축.

### [Phase 6] Integrity & Stability (신뢰성 강화) - ✅ 완료
- **feat/schema-validation**: `information_schema`를 통한 테이블/컬럼 존재 여부 사전 체크.
- **feat/debug-mode**: `--verbose` 플래그를 통한 분석 중간 과정 상세 노출.
- **feat/unit-testing**: 리스크 엔진 및 파서에 대한 시나리오 기반 유닛 테스트 강화.

---

## 2. 향후 릴리즈 계획 (Post v3.1 - Refactoring & Expansion)

### [Phase 7] Architectural Refactoring (진행 예정)
- **refactor/cmd-logic**: `cmd/` 내의 비즈니스 로직을 `internal/service` 등으로 분리.
- **refactor/error-handling**: 전용 에러 타입 정의 및 에러 전파 체계 개선.
- **refactor/di**: 인터페이스 기반 의존성 주입(Dependency Injection) 적용.

### [Phase 8] CI/CD & Ecosystem
- **feat/github-actions**: GitHub Actions 워크플로우 템플릿 및 배포 가이드.
- **feat/api-mode**: 에이전트를 HTTP API 서버로 전환하여 원격 분석 지원.
