# 🧪 MigraGuard v3.4 Simulation & Validation Guide

본 문서는 실환경 모사 부하 생성기를 사용하여 MigraGuard의 리스크 탐지 능력을 검증하는 절차를 안내합니다.

---

## 1. Environment Setup

### [1-1] Clean Start
데이터 스키마 및 볼륨 초기화를 위해 다음 명령을 수행합니다.

```bash
docker compose down -v
docker compose up -d --build
```

### [1-2] Data Volume Verification
`init-db.sql`에 의해 `orders` 테이블에 10만 건 이상의 데이터가 정상적으로 적재되었는지 확인합니다.

---

## 2. Dynamic Traffic Simulation

### [2-1] Normal Load (Steady State)
기본적으로 `docker-compose` 실행 시 15명의 워커가 평상시 트래픽을 생성합니다.

### [2-2] Peak Load (Flash Sale)
특수한 이벤트 상황을 재현하려면 워커 수를 늘려 별도로 실행합니다.

```bash
# 기존 제네레이터 중단 후 고부하 모드 실행
docker compose stop load-generator
docker compose run -d --name load-gen-peak load-generator --conns 50 --profile flash-sale
```

---

## 3. Risk Analysis Validation

트래픽이 활발한 상태에서 다양한 DDL 시나리오를 분석합니다.

### [3-1] Heavy Migration Test (Danger)
```bash
docker compose run --rm analyze-shell analyze /app/code/migrations/007_danger_rewrite_type.sql \
  --db "postgres://user:pass@db:5432/target_db" \
  --sqlite "/app/data/migraguard.db"
```

### [3-2] Safe Migration Test (Safe)
```bash
docker compose run --rm analyze-shell analyze /app/code/migrations/004_safe_add_column.sql \
  --db "postgres://user:pass@db:5432/target_db" \
  --sqlite "/app/data/migraguard.db"
```

---

## 4. Verification Checklist

1. **Traffic Intensity**: 리포트의 `Avg(1h)` 및 `Peak(24h)` 지표가 부하 상황에 따라 수백 TPS 단위로 정상 출력되는가?
2. **Dynamic Gating**: 동일한 DDL이라도 트래픽이 적을 때(Safe)와 많을 때(Danger) 리스크 등급이 동적으로 변하는가?
3. **Report Accuracy**: 배포 팁(Tip)에서 추천 시간대로 변경 시 예상되는 트래픽 감소량이 백분율(%)로 정확히 계산되는가?

## 5. Load Scenario Matrix (v3.4)

| 시나리오 | 목표 TPS | 동시성 | 주요 탐지 항목 |
| :--- | :--- | :--- | :--- |
| **Idle** | 0 | 0 | 메타데이터 기반 최소 리스크 |
| **Steady** | ~100 | 15 | 트래픽 가중치 반영 여부 |
| **Flash-Sale** | 500+ | 50 | 대기열 폭증 및 배포 차단 |

---
*Last Updated: 2026-04-06 (v3.4 Simulation Guide)*
