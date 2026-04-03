# 🛡️ MigraGuard 프로젝트 진행 상황 (v3.1 아키텍처 대전환)

## 📅 마지막 업데이트: 2026-04-03
- **현재 상태:** **v3.1 에이전트-CLI 분리 모델로의 아키텍처 재설계 및 명세서 수립 완료**
- **핵심 변경 사항:** 분석 시점에만 임시로 수집하던 방식(v3.0)에서, 상시 수집 에이전트(Agent)와 즉각 분석 도구(CLI)가 공존하는 구조로 전환.

## 🛠️ v3.1 전환 로드맵
1. **[Phase 1] 커맨드 분리**: `cmd/migraguard/agent.go`와 `analyze.go`로 명령 체계 이원화.
2. **[Phase 2] 에이전트 고도화**: `Collector`의 상시 실행 안정성 확보 및 데이터 보존 정책(Retention) 구현.
3. **[Phase 3] 리스크 엔진 개편**: 3초 대기(`time.Sleep`) 로직을 제거하고 SQLite 시계열 쿼리(AVG, MAX) 기반으로 변경.
4. **[Phase 4] 인프라 정의**: Docker Compose를 이용한 Agent(상시)와 CLI(트리거) 간의 데이터 공유(Volume) 가이드라인 작성.

## ✅ 기존 완료 항목 (유지 및 재활용)
- `internal/parser/ast.go`: SQL 파싱 및 락 레벨 식별 로직은 그대로 유지.
- `internal/db/activity.go`: PostgreSQL 메트릭 수집 기본 로직 유지.
- `internal/engine/risk.go`: 핵심 큐잉 모델 수식 유지 (입력값 $\lambda$의 출처만 변경).

## 🚀 향후 과제 (Next Steps)
- `migraguard agent` 커맨드 구현 및 백그라운드 무한 루프 로직 작성.
- SQLite 시계열 데이터 관리용 쿼리 최적화.
- Docker 볼륨 공유를 통한 테스트 환경 구축.
