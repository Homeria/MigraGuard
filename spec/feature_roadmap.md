# MigraGuard 기능 분할 및 브랜치 전략 (Feature Roadmap v3.1)

## 1. 기능 분할 전략 (Feature Breakdown)

### [Phase 1~7] Architecture & Foundation - ✅ 완료
- **Phase 1~6**: 에이전트 독립화, 리스크 엔진 고도화, 설정 파일, 도커 환경, 테스트 강화.
- **Phase 7 (Refactoring)**: 서비스 레이어 분리, DI 적용, 에러/리포터 추상화 완료.

---

## 2. 향후 로드맵 (Phase 8~9 - Ecosystem & Advanced)

리팩토링된 아키텍처를 기반으로 실제 운영 환경의 편의성을 극대화합니다.

### [Phase 8] CI/CD & GitHub Integration (Next)
- **Branch**: `feat/github-actions`
- **Goal**: GitHub PR에 분석 결과를 자동으로 게시하고 배포를 제어하는 생태계 구축.
- **Tasks**:
  - GitHub Actions용 `entrypoint.sh` 및 전용 Docker 이미지 배포.
  - `MarkdownReporter` 결과물을 GitHub Comment로 등록하는 가이드 작성.

### [Phase 9] Agent API Mode (v4.0 Preparation)
- **Branch**: `feat/api-mode`
- **Goal**: SQLite 파일 공유 방식의 한계를 넘어 네트워크 기반의 원격 분석 지원.
- **Tasks**:
  - Agent 내 HTTP 서버(Gin/Echo) 탑재.
  - `/metrics/:table`, `/risk/analyze` 등의 엔드포인트 구현.
  - CLI가 SQLite 직접 조회 대신 API를 호출하도록 전환.

### [Phase 10] ML-based Traffic Prediction
- **Branch**: `feat/traffic-ml`
- **Goal**: 단순 통계(Average, Peak)를 넘어 시계열 예측 모델을 통한 리스크 산출.

---

## 3. 향후 확장 계획 (Post v3.1)

### [Phase 8] CI/CD & Ecosystem
- **feat/github-actions**: GitHub Actions 워크플로우 템플릿 및 배포 가이드.
- **feat/api-mode**: 에이전트를 HTTP API 서버로 전환하여 원격 분석 지원.
