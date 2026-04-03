# 🛡️ MigraGuard 프로젝트 진행 상황 (v3.1 에이전트 모델 완성)

본 문서는 에이전트-CLI 분리 모델(v3.1)로의 아키텍처 대전환 및 구현 완료 상태를 기록합니다.

## 📅 마지막 업데이트: 2026-04-03
- **현재 상태:** **Phase 1~4 아키텍처 전면 재설계 및 구현 완료 (v3.1 에이전트 모델)**
- **핵심 성과:** 
  - **`migraguard agent` 구현**: 운영 DB 트래픽을 24/7 상시 수집하여 SQLite에 시계열 데이터를 구축하는 독립 에이전트 완성.
  - **`migraguard analyze` 즉각화**: 분석 시 3초간 대기하던 임시 로직을 제거하고, 에이전트가 쌓은 데이터를 즉각 활용하는 분석 체계 구축.
  - **데이터 관리 정책 도입**: 오래된 데이터를 삭제(Purge)하고 DB를 최적화(VACUUM)하는 Retention Policy 구현.
  - **명세서 최신화**: 모든 설계 문서(`spec/*.md`)를 v3.1 에이전트-CLI 모델로 현행화.

## ✅ 지금까지 완료된 작업 (v3.1)
1. **에이전트 독립화 (Phase 2 고도화)**:
   - `cmd/migraguard/agent.go`: 독립 실행형 에이전트 커맨드 신설.
   - `internal/db/collector.go`: 수집 루프 내 데이터 보존 정책(Default 7일) 및 에러 처리 강화.
   - `internal/db/activity.go`: `PurgeOldSnapshots` 메서드 추가 및 SQLite 최적화 로직 구현.
2. **분석 CLI 고속화 (Phase 4 고도화)**:
   - `cmd/migraguard/analyze.go`: 임시 수집기(`time.Sleep`) 제거 및 즉각 분석 로직 전환.
3. **리스크 엔진 연동 (Phase 3 유지)**:
   - `internal/engine/risk.go`: SQLite 델타 TPS 데이터를 최우선으로 사용하여 분석 정밀도 유지.

## 🛠️ 기술 사양 (v3.1 에이전트 모델)
- **실행 구조**: `Agent (상시 수집)` ↔ `SQLite (공유 볼륨)` ↔ `Analyze CLI (즉각 분석)`
- **데이터 보존**: 7일 경과 데이터 자동 삭제 및 주기적 VACUUM 수행.
- **분석 지표**: 실시간 델타 TPS($\lambda$)를 기반으로 한 큐잉 모델($C_{peak}$) 적용.

## 🚀 향후 과제 (Next Steps)
1. **상수 외부화 (Configuration)**:
   - `DiskIO`, `CMax`, `MuMax` 등 리스크 엔진 상수를 `migraguard.yaml` 설정 파일로 분리하여 환경별 유연성 확보.
2. **도커 환경 검증 (Validation)**:
   - `docker-compose`를 이용해 Agent 컨테이너와 CLI 실행 환경 간의 데이터 공유(Volume) 및 실제 작동 여부 테스트.
3. **Markdown 리포터 개발**:
   - CI 환경에서 PR 코멘트로 분석 결과를 남기기 위한 마크다운 생성기 구현.
4. **스키마 정합성 검증**:
   - 분석 중인 SQL의 테이블/컬럼이 실제 DB 스키마(`information_schema`)와 일치하는지 대조 로직 추가.

---
**세션 종료:** v3.1 에이전트 모델로의 대전환을 통해 프로토타입 수준을 넘어 실질적인 DevSecOps 도구로서의 기반을 완성했습니다. 수고하셨습니다!
