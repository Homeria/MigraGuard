# pgbench 실험 환경

이 디렉터리는 MigraGuard의 실측 비교 실험에 사용할 재현 가능한 PostgreSQL 작업 부하를 보관한다.

## 기준 부하 실행

PowerShell에서 다음 명령을 실행한다.

```powershell
docker compose up -d db benchmark
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\ps1\benchmark\run-baseline.ps1 -Workload orders_mixed -TargetTps 50 -DurationSeconds 30
```

`-ExecutionPolicy Bypass`는 현재 프로세스에서만 적용되며 시스템 실행 정책을 변경하지 않는다.

실행할 때마다 `benchmark_run` 데이터베이스를 새로 만들고 동일한 초기 데이터를 적재한다. 결과는 `experiments/benchmark/results/<실행 이름>`에 저장된다.

짧은 반복 실행에서 데이터 초기화를 생략하려면 `-SkipReset`을 사용한다. DDL을 적용한 실험을 서로 비교할 때는 이 옵션을 사용하지 않는다.

## 부하 중 DDL 실행

다음 명령은 기준 데이터를 초기화하고, 작업 부하를 시작한 지 5초 뒤 DDL을 실행한다. Squawk 결과, DDL 수행 시간, 작업 부하 지연, 커넥션과 잠금 대기를 하나의 실행 디렉터리에 저장한다.

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\ps1\benchmark\run-ddl-case.ps1 `
  -DdlPath .\experiments\ddl\022_safe_add_orders_note.sql `
  -Workload orders_mixed -TargetTps 50 -DurationSeconds 30 -WarmupSeconds 5
```

DDL과 작업 부하의 대상 테이블을 맞춰야 한다. `orders` DDL은 `orders_mixed`, `order_event_logs` DDL은 `logs_write`, `inventory_stocks` DDL은 `stock_update`, `account_balances` DDL은 `balance_update`와 조합한다.

## 작업 부하

| 이름 | 대상 | 주된 동작 |
|---|---|---|
| `orders_mixed` | `orders` | 주문 조회와 갱신 |
| `logs_write` | `order_event_logs` | 이벤트 삽입과 최근 이벤트 조회 |
| `stock_update` | `inventory_stocks` | 재고 행 갱신과 조회 |
| `balance_update` | `account_balances` | 잔액 버전 갱신과 조회 |

`TargetTps`는 pgbench의 목표 트랜잭션 처리율이며, 서버가 이를 처리하지 못하면 실제 TPS와 지연이 달라질 수 있다. `db_metrics.csv`에는 1초 간격의 커넥션 수와 잠금 대기 수가, `pgbench_summary.txt`와 `pgbench_log.*`에는 처리율과 개별 트랜잭션 지연이 기록된다. `summary_metrics.json`에는 실제 TPS와 P50, P95, P99, 최대 지연이 요약된다.
