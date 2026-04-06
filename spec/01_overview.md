# 🛡️ MigraGuard Project Overview (v3.4 Refined)

## 1. Project Definition (What is MigraGuard?)
**MigraGuard**는 PostgreSQL의 스키마 변경(DDL) 시 발생할 수 있는 서비스 장애 리스크를 사전에 탐지하는 **트래픽 인지형(Traffic-Aware) DB 가드**입니다. 24/7 백그라운드 수집 **Agent**와 실환경 모사 **Load Generator**를 통해 DDL 배포의 안전성을 정밀 검증합니다.

## 2. Core Value
- **Precise Load Tracking:** 주기별 실시간 부하(Delta) 측정을 통한 정확한 TPS 산출.
- **Service Mimicry:** 24시간 트래픽 곡선과 비즈니스 시나리오를 반영한 고충실도 시뮬레이션.
- **Conservative Risk Model:** 트래픽 강도에 따라 유동적으로 변화하는 5단계 위험도 평가 엔진.
- **Automated Gatekeeping:** 위험한 DDL 감지 시 배포 프로세스 즉시 차단.

## 3. Key Features (v3.4)
1. **High-Fidelity Load Generator:**
   - 이커머스 도메인(`orders`, `products`) 기반의 복잡한 트랜잭션 부하 생성.
   - 시간대별 가중치(Sine Wave)를 적용한 현실적인 트래픽 패턴 재현.
2. **Refined Background Agent:**
   - PostgreSQL 14+ 호환성 확보 및 테이블별 정밀 지표 수집 기능 탑재.
3. **Dynamic Risk Evaluation Engine:**
   - 실시간 트래픽 가중치($\lambda_{final}$)를 통한 보수적 리스크 등급 산출.

## 4. Feature Roadmap & Milestones

### ✅ v3.4 Simulation & Empirical Validation (Current State)
- **Load Generator Implementation**: Go 기반의 커스텀 부하 생성기 및 CLI 통합 완료.
- **Risk Engine Verification**: 10만 건 이상의 데이터 환경에서 DDL 위험도 실증 완료.

### 🚀 Future Roadmap (Phase 9 ~ 10)
1. **[v3.5] CI/CD Pipeline Integration**:
   - `feat/github-integration`: GitHub Actions 연동 및 PR 자동 리포팅.
2. **[v4.0] AI-Agent MCP Server**:
   - `feat/api-mode`: AI 에이전트용 DB 인텔리전스 제공 MCP 서버 확장.

---
*Last Updated: 2026-04-06 (v3.4 Finalized)*
