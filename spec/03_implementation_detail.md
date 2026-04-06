# ⚙️ MigraGuard v3.4 Implementation Details

본 문서는 v3.4에서 구현된 **고충실도 부하 생성기**와 **리스크 엔진 최적화**의 상세 기술 명세를 다룹니다.

---

## 1. Load Generator Engine (`internal/simulation/`)

- **Objective:** 실제 이커머스 서비스의 트래픽 패턴을 모사하여 DB 락 경합 및 성능 저하 상황을 재현.
- **Worker Pool Architecture:**
  - `Concurrency` 설정에 따른 다중 고루틴 워커 할당.
  - 각 워커는 독립적인 세션으로 DB 요청을 수행하여 동시성 극대화.
- **24-Hour Traffic Curve (Sine Wave):**
  - $\text{Intensity} = \frac{\sin(\frac{\pi \times (\text{hour}-9)}{12}) + 1.5}{2.5}$
  - 현재 시스템 시간을 기반으로 부하 강도를 0.2 ~ 1.0 사이로 자동 조절하여 현실적인 일간 트래픽 곡선 생성.
- **Business Scenarios:**
  - **Browse**: 사용자 및 상품 목록 조회 (SELECT 중심).
  - **Order**: 주문 생성, 상세 기록, 상품 재고 차감을 포함한 트랜잭션 수행 (쓰기 및 락 경합 중심).

## 2. Risk Evaluation Sensitivity Tuning

- **Dynamic Thresholding:** 
  - 테스트 환경 실증을 위해 `CMax`(시스템 최대 허용 연결 수)를 하향 조정(예: 100)하여 민감한 위험 탐지 보장.
- **Conservative Weighting:**
  - 리스크 점수 산출 시 $\max(\text{Current}, \text{Avg}_{1h} \times 1.2, \text{Peak}_{24h} \times 0.8)$를 적용하여 불확실성이 높은 상황에서 안전하게 배포 차단.

## 3. PostgreSQL Compatibility (`internal/db/postgres_adapter.go`)

- **PG 14+ Stats Support:**
  - `pg_stat_statements`에서 삭제된 `stats_reset` 컬럼 대신 `pg_stat_statements_info` 뷰를 참조하도록 TPS 계산 쿼리 개선.
- **Case-Insensitive Table Matching:**
  - 쿼리 텍스트 매칭 시 대소문자 구분을 없애 실시간 지표 수집의 누락 방지.

---

## Technical Summary: Simulation Workflow
1. **Load Generator**가 설정된 프로파일에 따라 `orders` 테이블 등에 지속적인 트랜잭션 부하 유발.
2. **Agent**가 PostgreSQL 15의 통계 뷰를 분석하여 테이블별 정밀 지표 수집.
3. **Analyze CLI**가 수집된 지표를 바탕으로 현재 트래픽이 DDL 배포에 미칠 악영향을 수치화.
4. 10만 건 이상의 대형 테이블 변경 시 **Danger** 등급을 부여하여 실제 서비스 장애 예방 능력 검증.

---
*Last Updated: 2026-04-06 (v3.4 Simulation Detail)*
