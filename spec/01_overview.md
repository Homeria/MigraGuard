# 🛡️ MigraGuard Project Overview (v3.3 Refined)

## 1. Project Definition (What is MigraGuard?)
**MigraGuard**는 PostgreSQL 데이터베이스의 스카마 변경(DDL) 시 발생할 수 있는 서비스 장애 리스크를 사전에 탐지하는 **트래픽 인지형(Traffic-Aware) DB 가드**입니다. 24/7 백그라운드에서 운영 트래픽을 수집하는 **Agent**와 배포 전 SQL 리스크를 정밀 분석하는 **Analyze CLI**가 한 쌍으로 작동합니다.

## 2. Core Value
- **Precise Load Tracking:** `pg_stat_statements`의 누적치가 아닌 주기별 실시간 부하(Delta)를 측정하여 정확도 극대화.
- **Dynamic Risk Score:** 정적 SQL 분석과 실시간 피크 트래픽 가중치를 결합한 5단계 진단 알고리즘.
- **Immediate Feedback:** 공유 SQLite 저장소를 통해 배포 시점에 추가적인 DB 부하 없이 즉각적인 리스크 등급 산출.
- **Service Simulation:** 실제 서비스 시나리오를 모사한 부하 테스트를 통해 리스크 판단 능력 검증.

## 3. Key Features (v3.3)
1. **Refined Background Agent:**
   - 구간별 Delta 수집 및 `original_pg_stat_statements` 영속성 관리를 통한 데이터 정밀도 확보.
2. **Modular DB Adapter Layer:**
   - PostgreSQL/SQLite 어댑터를 모델, 저장소, 분석기로 세분화하여 독립적인 유지보수성 확보.
3. **5-Step Risk Evaluation Engine:**
   - $T_{ddl}$ (스키마 분석), $T_{block}$ (락 경합), $C_{peak}$ (피크 동시성), $T_{rec}$ (회복 비용), 최종 리스크 점수 산출.

## 4. Feature Roadmap & Milestones

### ✅ v3.3 Refined & Refactored (Current State)
- **Delta Collection Implementation**: 구간별 부하량 추출 및 상태 보존 로직 완료.
- **DB Layer Modularization**: 도메인 모델 분리 및 어댑터 레이어 모듈화 완료.

### 🚀 Future Roadmap (v3.4 ~ v4.0)
1. **[v3.4] Real-world Service Simulator**:
   - `feat/load-generator`: 단순 `pgbench`가 아닌 실서비스 비즈니스 프로파일을 모사한 부하 생성기 도입.
2. **[v3.5] CI/CD Pipeline Integration**:
   - `feat/github-integration`: PR 기반 자동 리포트 생성 및 배포 차단 파이프라인.
3. **[v4.0] AI-Agent MCP Server**:
   - `feat/api-mode`: AI 코딩 에이전트가 실시간 DB 상태를 조회하고 가이드를 줄 수 있는 MCP 서버 확장.

---
*Last Updated: 2026-04-06 (v3.3 Finalized)*
