# 🛡️ MigraGuard 프로젝트 진행 상황 (v3.3 정밀 수집 설계 진입)

본 문서는 v3.2 아키텍처 완성과 더불어, 데이터 수집 정밀도를 극대화하기 위한 v3.3 설계 단계를 기록합니다.

## 📅 마지막 업데이트: 2026-04-05 (v3.3 설계 중)
- **현재 상태:** **v3.3 정밀 데이터 수집(Delta Collection) 설계 단계**
- **핵심 성과:** 
  - **v3.2 완료**: Docker 기반 이중 프로세스 아키텍처 및 실환경 검증 인프라 구축 완료.
  - **Delta Logic Design**: `pg_stat_statements`의 누적치를 구간별 차이값(Delta)으로 변환하는 저장 로직 설계 완료.
  - **Persistence Strategy**: 메모리가 아닌 SQLite 내 `original_pg_stat_statements` 테이블을 활용한 데이터 보존 전략 수립.

## ✅ 완료된 작업 (Milestones)
1. **v3.2 아키텍처 고도화**: 도메인 서비스 분리(SOA) 및 Docker Volume을 통한 데이터 공유 구조 완성.
2. **Infra Automation**: `init-db.sql` 및 영속성 볼륨 설정을 통한 원클릭 환경 구축.
3. **Spec Refactoring**: 12개의 문서를 4개의 핵심 설계 문서로 통합 및 리팩토링 완료.

## 🚀 향후 로드맵 (Phase 8~10 전략)
1. **`feat/delta-collection` (v3.3 핵심)**:
   - **Cumulative to Delta**: `pg_stat_statements`의 누적 데이터에서 수집 주기 사이의 실제 부하량(Delta)을 추출하는 로직 구현.
   - **Original Stats Tracking**: SQLite에 마지막 원본 수집 데이터를 별도로 보관하여 에이전트 재시작 후에도 정밀한 차이값 계산 보장.
2. **`feat/github-integration` (v3.4 예정)**:
   - GitHub Actions 워크플로우 및 PR 코멘트 봇 연동을 통한 CI/CD 통합.
3. **`feat/load-generator`**:
   - 실환경 시뮬레이션을 위한 커스텀 트래픽 생성 툴 개발.

---
**세션 종료:** v3.3의 핵심인 정밀 수집 로직 설계가 완료되었습니다. 이제 누적값이 아닌 실제 구간별 부하량을 기반으로 한 더욱 정확한 리스크 분석이 가능해질 것입니다.
