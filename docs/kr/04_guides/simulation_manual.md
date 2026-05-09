# 🧪 MigraGuard 시나리오 검증 가이드 (Simulation Guide)

본 문서는 실환경 모사 부하 생성기(Load Generator)를 사용하여 MigraGuard의 리스크 탐지 능력을 검증하는 절차를 안내합니다.

---

## 1. 환경 초기화
```bash
docker compose down -v
docker compose up -d --build
```

## 2. 부하 시뮬레이션
- **일반 부하**: `docker-compose` 실행 시 자동 시작 (워커 15명).
- **피크 부하 (Flash Sale)**:
  ```bash
  docker compose run -d --name load-gen-peak load-generator --conns 50 --profile flash-sale
  ```

## 3. 리스크 분석 검증
- **테이블 재작성 (Danger)**: `007_danger_rewrite_type.sql` 분석.
- **메타데이터 변경 (Safe)**: `004_safe_add_column.sql` 분석.

## 4. 검증 체크리스트
- 리포트의 `Avg(1h)` 및 `Peak(24h)` 지표가 실제 상황을 반영하는가?
- 트래픽 부하에 따라 동일 DDL의 등급이 유동적으로 변하는가?
- 추천된 골든 타임이 합리적인가?
