# 🧪 MigraGuard 부하 시뮬레이션 및 검증 가이드

본 문서는 실제 PostgreSQL 환경에서 부하를 생성하고, MigraGuard가 이를 어떻게 수집하고 분석하는지 검증하는 절차를 설명합니다.

---

## 1. 환경 준비 (Environment Setup)

### [1-1] Docker 인프라 가동
프로젝트 루트의 `docker-compose.yml`을 사용하여 운영 DB와 에이전트를 띄웁니다.
```bash
docker-compose up -d
```
- **PostgreSQL**: `localhost:5432` (user: postgres, db: postgres)
- **Agent**: 백그라운드에서 지표 수집 시작.

### [1-2] 테스트용 데이터 생성 (pgbench 활용)
PostgreSQL 기본 부하 테스트 도구인 `pgbench`를 사용하여 초기 데이터를 생성합니다.
```bash
# 초기화 (Scale factor 10 -> 약 100만건 데이터)
docker exec -it postgres pgbench -i -s 10 postgres
```

---

## 2. 부하 생성 (Load Generation)

### [2-1] 일반 트래픽 시뮬레이션 (Baseline)
낮은 수준의 부하를 주어 '정상 상태' 지표를 쌓습니다.
```bash
# 10명의 사용자가 5분간 지속적으로 쿼리 실행
docker exec -it postgres pgbench -c 10 -T 300 postgres
```

### [2-2] 피크 트래픽 시뮬레이션 (Peak Load)
에이전트가 `PeakTPS24h`를 기록할 수 있도록 높은 부하를 순간적으로 줍니다.
```bash
# 50명의 사용자가 1분간 매우 빠르게 쿼리 실행
docker exec -it postgres pgbench -c 50 -j 4 -T 60 postgres
```

---

## 3. 에이전트 수집 확인 (Agent Verification)

에이전트가 SQLite에 지표를 잘 저장하고 있는지 쿼리해 봅니다.
```bash
# SQLite 접속
sqlite3 data/migraguard.db

# 최근 수집된 TPS 확인
sqlite> SELECT timestamp, table_name, tps FROM table_metrics ORDER BY timestamp DESC LIMIT 10;
```

---

## 4. DDL 리스크 분석 테스트 (Analyze Validation)

부하가 걸려 있는 상태(또는 부하 직후)에서 분석을 수행하여 $RiskScore$ 변화를 확인합니다.

### [4-1] 테스트용 SQL 작성 (`test_migration.sql`)
```sql
ALTER TABLE pgbench_accounts ALTER COLUMN abalance TYPE BIGINT;
```

### [4-2] 분석 실행
```bash
# 부하가 없을 때 실행
./migraguard analyze test_migration.sql --db "..." --sqlite "..."

# pgbench 부하를 주는 도중에 실행 (RiskScore 상승 확인)
./migraguard analyze test_migration.sql --db "..." --sqlite "..."
```

---

## 5. 체크리스트 (Success Criteria)

1.  **지표 연속성**: `table_metrics` 테이블에 1분 단위로 끊김 없이 데이터가 쌓이는가?
2.  **가중치 로직**: `Analyze` 실행 시 SQLite에서 가져온 `AvgTPS1h`와 `PeakTPS24h`가 리포트에 포함되는가?
3.  **게이트키핑**: $RiskScore$가 90%를 넘을 때 실제로 Exit Code 1이 발생하는가?
4.  **공간 관리**: `PurgeOldSnapshots` 로직이 실행되어 SQLite 파일 크기가 일정하게 유지되는가?
