# 🛡️ MigraGuard 프로젝트 진행 상황 (Session Handover)

본 문서는 다른 세션에서 작업을 이어받기 위한 가이드 및 현재 상태 기록입니다.

## 📅 마지막 업데이트: 2026-04-02
- **현재 브랜치:** `develop` (Phase 2 완료 후 통합됨)
- **핵심 목표:** DB 마이그레이션 시 락 경합 방지를 위한 CLI 게이트키퍼 구축

## 🛠️ 코딩 규칙 (Coding Standards) - 중요!
모든 코드 작성 시 다음 규칙을 엄격히 준수해야 합니다.
1. **주석 언어 분리**:
   - **영문 (English)**: 함수, 타입, 구조체 필드 등 외부로 노출되는 공식 문서화 주석.
   - **한글 (Korean)**: 코드 내부의 상세 구현 설명, 로직 힌트, `TODO` 주석.
2. **CGO 환경**:
   - `pg_query_go` 및 `go-sqlite3` 라이브러리를 사용하므로, 빌드 및 테스트 환경에 `gcc`가 반드시 설치되어 있어야 함.
3. **보수적 설계**:
   - 프로젝트 정체성이 확립되는 단계이므로, 과도한 보일러플레이트 생성보다는 명세(`spec/`)에 충실한 핵심 로직 구현에 집중할 것.

## ✅ 지금까지 완료된 작업
1. **Phase 1: Core Parser 구현 완료**:
   - SQL을 AST로 변환하고 테이블 명, 컬럼 명, PostgreSQL 락 레벨을 추출하는 엔진 구축.
2. **Phase 2: Data Foundation 구현 완료**:
   - `internal/db/connection.go`: `pgxpool`을 이용한 PostgreSQL 연결 관리.
   - `internal/db/workload.go`: `pg_stat_statements` 기반 워크로드 스냅샷 수집.
   - `internal/db/activity.go`: SQLite WAL 모드 기반 로컬 시계열 저장소 및 데이터 분석 레이어.
   - `internal/db/collector.go`: 고루틴 Ticker를 이용한 백그라운드 데이터 수집 엔진.
3. **명세 고도화**:
   - `spec/feature_roadmap.md`: 전체 페이즈별 세부 체크포인트 수립 및 업데이트.

## 🚀 다음 세션에서 수행할 작업 (Next Steps)
1. **Phase 3: Risk Engine 시작**:
   - `internal/engine/risk.go`: `Traffic * Lock * Wait` 기반의 정량적 Risk Score 산출 알고리즘 구현.
2. **알고리즘 정교화**:
   - SQLite에 저장된 과거 트래픽(Average/Max Calls)을 Risk Score에 반영하는 로직 설계.
3. **Safe Window 추천**:
   - 트래픽 밀도가 낮은 최적의 배포 시간대 추천 기능 개발.

---
**세션 연결 가이드:** 다음 세션 시작 시 "GEMINI.md와 spec 폴더를 읽고 프로젝트의 현재 상태와 주석 규칙을 파악해줘"라고 명령하세요.
