# 🧪 MigraGuard v3.3 Simulation & Validation Guide

본 문서는 MigraGuard의 리스크 엔진을 검증하기 위해 실환경과 유사한 트래픽 부하를 생성하고 분석하는 절차를 설명합니다.

---

## 1. Environment Setup

### [1-1] Docker Infrastructure
`docker-compose.yml`을 통해 운영 DB, 수집 에이전트, 부하 생성기(Load Generator)를 통합 실행합니다.

```bash
docker-compose down -v  # 기존 데이터 초기화
docker-compose up -d --build
```

### [1-2] Database Initial State
`init-db.sql`에 정의된 비즈니스 테이블(`users`, `orders`, `products` 등)이 자동 생성되었는지 확인합니다.

---

## 2. Real-world Traffic Simulation

단순 `pgbench` 대신 시나리오 기반 시뮬레이션을 수행합니다.

### [2-1] Simulation Profiles
시뮬레이터는 설정된 비즈니스 프로파일에 따라 쿼리 믹스를 생성합니다.

- **Profile: Normal Operation**
  - Read:Write = 90:10
  - Concurrency: 5~10
- **Profile: Marketing Event (Peak)**
  - Read:Write = 50:50
  - Concurrency: 50~100
  - 쓰기 집중 테이블: `orders`, `inventory`

### [2-2] Load Generation (Execution)
```bash
# 시뮬레이터 실행 (예: 500 TPS 목표 부하)
docker compose run load-generator --profile event --target-tps 500
```

---

## 3. Risk Analysis Validation

부하가 발생 중인 상황에서 마이그레이션 SQL을 분석하여 리스크 등급의 변화를 확인합니다.

### [3-1] Test Scenario: Heavy Alter
```sql
-- rewrite_type.sql
ALTER TABLE orders ALTER COLUMN status TYPE TEXT;
```

### [3-2] Step-by-Step Validation
1. **정상 부하 (10 TPS)**: 분석 결과 `Safe` 판정 확인.
2. **이벤트 부하 (200 TPS)**: 동일 SQL 분석 결과 `Warning` 판정 확인.
3. **극한 부하 (500 TPS)**: 동일 SQL 분석 결과 `Danger` 판정 및 배포 차단(Exit Code 1) 확인.

---

## 4. Verification Checklist

1. **Delta Accuracy**: `workload_snapshots` 테이블의 `calls` 값이 수집 간격 동안의 실제 실행 횟수와 일치하는가?
2. **Peak Sensitivity**: 24시간 피크 트래픽이 `RiskScore`에 적절한 가중치($\times 0.8$)로 반영되는가?
3. **Continuity**: 에이전트 재시작 후에도 델타 계산 결과가 튀지 않고 안정적인가?

## 5. Summary Table: Validation Scenarios

| 부하 시나리오 | TPS | 동시성 (Conn) | 예상 리스크 등급 |
| :--- | :--- | :--- | :--- |
| **Idle** | < 1 | 1 | **Safe** |
| **Steady State** | 50 | 10 | **Safe** |
| **High Traffic** | 200 | 50 | **Warning** |
| **Heavy Load + Rewrite** | 500+ | 100+ | **Danger** |

---
*Last Updated: 2026-04-06 (v3.3 Simulation Update)*
