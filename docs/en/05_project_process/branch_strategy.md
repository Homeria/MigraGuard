# 🌿 Git Branch Strategy & Workflow

MigraGuard 프로젝트는 캡스톤 디자인의 체계적인 형상 관리를 위해 **Git Flow**를 기반으로 한 브랜치 전략을 채택합니다.

---

## 🏗️ 1. Branch Hierarchy (브랜치 계층 구조)

| Branch Category | Prefix | Target | Description |
| :--- | :--- | :--- | :--- |
| **Main** | `main` | - | 운영 환경 및 릴리스 가능한 최종 결과물 보관 |
| **Feature** | `feat/` | `main` | 신규 기능 개발 (예: `feat/adaptive-config`) |
| **Documentation**| `docs/` | `main` | 문서 작성 및 설계도 업데이트 (현재: `docs/capstone-requirements`) |
| **Bug Fix** | `fix/` | `main` | 기존 기능의 버그 수정 |
| **Refactor** | `refactor/`| `main` | 코드 성능 개선 및 구조 변경 (기능 변화 없음) |

---

## 🚀 2. Workflow (개발 흐름)

1.  **Issue Generation**: 개발할 기능이나 문서 작업을 이슈로 등록합니다.
2.  **Branch Checkout**: 작업 성격에 맞는 접두사를 사용하여 새 브랜치를 생성합니다.
    - 예: `git checkout -b feat/mcp-server`
3.  **Atomic Commits**: 한 커밋에 한 기능만 포함하도록 작고 명확하게 커밋합니다.
4.  **Pull Request (PR)**: 작업 완료 후 `main` 브랜치로 PR을 생성합니다.
5.  **Risk Analysis**: MigraGuard 자체 분석 결과(보고서)를 PR 코멘트에 첨부하여 검증합니다.
6.  **Merge**: 검토 후 병합하며, 병합 후에는 작업 브랜치를 삭제하여 관리 효율을 높입니다.

---

## 📝 3. Commit Convention (커밋 컨벤션)

- `feat:`: 새로운 기능 추가
- `fix:`: 버그 수정
- `docs:`: 문서 수정
- `style:`: 코드 포맷팅 (세미콜론 누락, 코드 변경 없음)
- `refactor:`: 코드 리팩토링
- `test:`: 테스트 코드 추가 및 수정
- `chore:`: 빌드 업무 수정, 패키지 매니저 수정 등

---

## 📅 4. Release Strategy

- 캡스톤 최종 발표 시점의 결과물은 `main` 브랜치에 태그(`v1.0.0`)를 부여하여 보관합니다.
- 각 Phase(v3.4, v3.5 등) 완료 시점마다 중간 릴리스를 기록합니다.

---

## 🗺️ 5. Branch Role Map (프로젝트 브랜치 역할 지도)

MigraGuard의 개발 단계별 주요 브랜치와 그 역할은 다음과 같습니다. (v3.5 기준)

### 🚀 Core Features (핵심 기능 개발)
- `feat/parser`: PostgreSQL SQL 파서(`pg_query_go`) 통합 및 AST 분석 로직 구현.
- `feat/db-adapter`: PostgreSQL(지표 조회) 및 SQLite(메트릭 저장)를 위한 추상화 레이어 구축.
- `feat/risk-engine`: 5단계 리스크 점수 산출 알고리즘 및 판정 규칙 구현.
- `feat/collector`: `pg_stat_statements` 기반의 지표 수집기 원형 개발.

### 🏗️ Architecture Refactoring (v3.2 아키텍처 고도화)
- `refactor/domain-service`: 비즈니스 로직을 서비스 레이어로 분리하여 결합도 감소.
- `refactor/dependency-injection`: 인터페이스 기반의 의존성 주입을 통해 테스트 용이성 확보.
- `refactor/error-management`: 구조화된 에러 핸들링 시스템 구축.
- `refactor/db-layer-modularization`: 영속성 계층을 모듈화하여 확장성 개선.

### 📈 Advanced Metrics (v3.3~v3.4 고도화)
- `feat/delta-collection`: 누적 데이터에서 시점별 변화량(Delta)을 추출하는 정밀 분석 로직 (v3.3).
- `feat/load-generator`: 이커머스 시나리오 기반의 고충실도 부하 생성기 구현 (v3.4).
- `feat/markdown-reporter`: CI/CD 연동을 위한 마크다운 형식의 리포트 자동 생성 기능.

### 📝 Current Phase (v3.5~ 설계 및 문서화)
- `docs/capstone-requirements`: (현재) 공학적 문서 체계 수립 및 요구사항 명세 고도화.
- `docs/restructure_and_stories`: 문서 구조 재편 및 사용자 스토리 작성.
