# 🧪 MigraGuard 시나리오 검증 및 운용 가이드 (Simulation Guide)

본 문서는 실환경 모사 부하 생성기를 사용하여 MigraGuard의 리스크 탐지 능력을 검증하고 실제 운용하는 절차를 안내합니다.

---

## 1. 환경 구축 및 초기화

### 1.1. 깨끗한 시작 (Clean Start)
데이터 스키마와 저장소 볼륨을 초기화하기 위해 다음 명령을 실행합니다.

```bash
docker compose down -v
docker compose up -d --build
```

### 1.2. 데이터 적재 확인
`init-db.sql`에 의해 `orders` 테이블에 약 10만 건 이상의 테스트용 데이터가 정상적으로 입력되었는지 확인합니다.

---

## 2. 동적 트래픽 시뮬레이션

### 2.1. 일반 부하 (Steady State)
`docker-compose` 실행 시 기본적으로 15명의 워커가 평상시 트래픽을 생성하며 에이전트가 이를 수집하기 시작합니다.

### 2.2. 피크 트래픽 재현 (Flash Sale)
특수한 이벤트 상황이나 대규모 장애 상황을 재현하려면 워커 수를 늘려 별도로 실행합니다.

```bash
# 기존 제네레이터 중단 후 고부하 모드 실행
docker compose stop load-generator
docker compose run -d --name load-gen-peak load-generator --conns 50 --profile flash-sale
```

---

## 3. 리스크 분석 검증 (Analyze)

트래픽이 활발하게 발생하는 도중에 다양한 DDL 시나리오를 분석하여 엔진의 반응을 확인합니다.

### 3.1. 고위험 마이그레이션 테스트 (Danger)
테이블 재작성을 유발하는 쿼리를 분석합니다.
```bash
docker compose run --rm analyze-shell analyze /app/code/migrations/007_danger_rewrite_type.sql \
  --db "postgres://user:pass@db:5432/target_db" \
  --sqlite "/app/data/migraguard.db"
```

### 3.2. 안전한 마이그레이션 테스트 (Safe)
단순 컬럼 추가 등 메타데이터만 변경하는 쿼리를 분석합니다.
```bash
docker compose run --rm analyze-shell analyze /app/code/migrations/004_safe_add_column.sql \
  --db "postgres://user:pass@db:5432/target_db" \
  --sqlite "/app/data/migraguard.db"
```

---

## 4. 검증 체크리스트 (Verification)

1. **트래픽 가시성**: 리포트의 `Avg(1h)` 및 `Peak(24h)` 지표가 실제 부하 상황에 맞게 수백 TPS 단위로 출력되는가?
2. **동적 게이트키핑**: 동일한 DDL이라도 트래픽이 적을 때(Safe)와 많을 때(Danger) 리스크 등급이 유동적으로 변하는가?
3. **추천 정확도**: 배포 팁(Tip)에서 제시하는 추천 시간대로 변경 시 예상되는 트래픽 감소량이 합리적으로 계산되는가?

---

## 5. 부하 시나리오 매트릭스

| 시나리오 | 목표 TPS | 동시성(Worker) | 주요 검증 항목 |
| :--- | :--- | :--- | :--- |
| **Idle** | 0 | 0 | 메타데이터 기반 최소 리스크 확인 |
| **Steady** | ~100 | 15 | 트래픽 가중치 반영 여부 확인 |
| **Flash-Sale** | 500+ | 50 | 대기열 폭증에 따른 배포 차단(Danger) 확인 |
