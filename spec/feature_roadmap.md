# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap v3.1)

## 1. 기능 분할 전략 (Feature Breakdown)

### [Phase 1~6] v3.1 Architecture & Stability - ✅ 완료
- **Phase 1~5**: 에이전트 독립화, 리스크 엔진 고도화, 설정 파일, 도커 환경 등 v3.1 핵심 아키텍처 완성.
- **Phase 6**: 스키마 정합성 검증, 유닛 테스트, 디버그 모드(Verbose) 구현 완료.

---

## 2. 대규모 리팩토링 로드맵 (Phase 7 - Architectural Refactoring)

현재의 기능 기반 구조를 서비스 중심의 유연한 구조로 재설계하기 위해 다음 순서로 작업을 진행합니다.

### [Phase 7-1] Domain Service Extraction (진행 예정)
- **Branch**: `refactor/domain-service`
- **Goal**: CLI 커맨드(`cmd/`)에 몰려있는 비즈니스 로직을 `internal/service`로 분리.
- **Tasks**: `AnalyzeService`, `AgentService` 구조체 신설 및 CLI 로직 이관.

### [Phase 7-2] Dependency Injection & Abstraction
- **Branch**: `refactor/dependency-injection`
- **Goal**: 서비스와 어댑터 간의 직접 의존성 제거 및 인터페이스 기반 주입 적용.
- **Tasks**: `DBAdapter` 인터페이스 정의 및 DI(Dependency Injection) 적용으로 테스트 격리성 확보.

### [Phase 7-3] Error Handling & Domain Errors
- **Branch**: `refactor/error-management`
- **Goal**: 하드코딩된 에러 메시지를 정적 타입 에러로 전환하고 에러 처리 일원화.
- **Tasks**: `internal/errors` 신설 및 도메인 전용 에러 코드 체계 구축.

### [Phase 7-4] Standardized Reporting
- **Branch**: `refactor/logging-reporter`
- **Goal**: 출력 로직을 `Reporter` 인터페이스로 추상화하여 확장성 확보.
- **Tasks**: `ConsoleReporter`와 `MarkdownReporter`를 통합 인터페이스로 관리.

---

## 3. 향후 확장 계획 (Post v3.1)

### [Phase 8] CI/CD & Ecosystem
- **feat/github-actions**: GitHub Actions 워크플로우 템플릿 및 배포 가이드.
- **feat/api-mode**: 에이전트를 HTTP API 서버로 전환하여 원격 분석 지원.
