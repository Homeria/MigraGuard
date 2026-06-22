# 실험 결과: CREATE INDEX CONCURRENTLY 기반 시간대 추천 검증

## 1. 실험 목적

본 실험의 목적은 MigraGuard가 forecast로 추천한 시간대가 실제 온라인 변경 작업을 수행하기에 적합한 시간대인지 확인하는 것이다. PostgreSQL 환경에서는 MySQL 기반 OSC 도구인 `gh-ost`, `pt-online-schema-change`를 직접 적용하기 어렵기 때문에, PostgreSQL이 공식 제공하는 DBMS 레벨 온라인 변경 기능인 `CREATE INDEX CONCURRENTLY`를 사용하였다.

검증 질문은 다음과 같다.

> MigraGuard가 추천한 낮은 위험 시간대에서 `CREATE INDEX CONCURRENTLY`를 실행했을 때, 다른 시간대보다 P99 응답 지연 증가, TPS 하락, 실패 트랜잭션 및 DDL 완료 시간이 더 작게 나타나는가?

## 2. 실험 조건

- DDL: `experiments/ddl/026_safe_concurrent_orders_composite_index.sql`
- 작업: `orders(status, created_at)` 복합 인덱스를 `CREATE INDEX CONCURRENTLY`로 생성
- 작업 부하: `orders_mixed`
- 대상 테이블: `orders`
- 데이터 규모: 기본 fixture + 추가 주문 데이터 1,000,000행
- 실행 시간: 각 조건 45초
- DDL 실행 시점: 작업 부하 시작 후 10초
- 결과 위치: `experiments/benchmark/results/concurrently_time_20260621_v2`

MigraGuard forecast에서 DDL-2는 02시를 Safe 추천 시간대로 제시하였다. 이를 실제 벤치마크에서는 낮은 부하 조건으로 재현하고, 중간 부하와 고부하 조건을 함께 비교하였다.

| 조건 | 의미 | Target TPS |
|---|---|---:|
| 추천 시간대 재현 | MigraGuard가 추천한 저부하 시간대 | 10 |
| 중간 부하 시간대 재현 | 일반 업무 시간대 수준 | 1000 |
| 고부하 시간대 재현 | MigraGuard가 Danger로 볼 가능성이 큰 시간대 | 5000 |

## 3. 측정 결과

| 조건 | Target TPS | 기준선 P99 | DDL 실행 중 P99 | P99 Ratio | Actual TPS Ratio | DDL 완료 시간 | 실패 트랜잭션 |
|---|---:|---:|---:|---:|---:|---:|---:|
| 추천 시간대 재현 | 10 | 3.73ms | 4.78ms | 1.28 | 0.86 | 683.43ms | 0 |
| 중간 부하 시간대 재현 | 1000 | 2.97ms | 2.94ms | 0.99 | 1.01 | 897.06ms | 0 |
| 고부하 시간대 재현 | 5000 | 8.25ms | 29.54ms | 3.58 | 1.00 | 874.96ms | 0 |

## 4. 해석

추천 시간대 재현 조건인 10 TPS에서는 DDL 완료 시간이 683.43ms로 가장 짧았고, 실패 트랜잭션도 발생하지 않았다. P99 응답 지연은 기준선 3.73ms에서 4.78ms로 증가했지만, 절대값 기준으로는 낮은 수준에 머물렀다.

중간 부하 조건인 1000 TPS에서도 결과는 안정적으로 나타났다. P99는 기준선 대비 거의 증가하지 않았고, Actual TPS도 기준선과 유사하였다. 따라서 이번 실험만으로는 MigraGuard 추천 시간대가 중간 부하 조건보다 항상 우수하다고 단정하기 어렵다.

반면 고부하 조건인 5000 TPS에서는 기준선 P99 8.25ms가 DDL 실행 중 29.54ms로 증가하여 P99 Ratio가 3.58로 나타났다. 실패 트랜잭션은 발생하지 않았지만, 온라인 변경 기능인 `CREATE INDEX CONCURRENTLY`를 사용하더라도 높은 부하 조건에서는 운영 쿼리 지연이 크게 증가할 수 있음을 확인하였다.

따라서 본 실험은 MigraGuard의 시간대 추천이 "유일한 최적 시간대"를 정확히 찾았음을 증명하기보다는, forecast가 고부하 시간대를 피하고 낮은 위험 후보를 제시하는 데 활용될 수 있음을 보여주는 예비 결과로 해석하는 것이 적절하다. 특히 추천 시간대와 중간 부하 시간대가 모두 안정적으로 나타난 점은 현재 RiskScore가 저부하 후보와 중간 부하 후보를 세밀하게 구분하기에는 추가 보정이 필요함을 의미한다.

## 5. 한계

- 본 실험은 MySQL 기반 OSC 도구를 직접 비교한 것이 아니라 PostgreSQL의 DBMS 레벨 온라인 변경 기능인 `CREATE INDEX CONCURRENTLY`를 사용한 예비 검증이다.
- `gh-ost`, `pt-online-schema-change`와 같은 OSC 도구의 chunk copy, throttle, pause, cut-over 시간을 직접 측정하지 않았다.
- 각 조건을 1회씩만 수행했으므로 반복 실험을 통한 통계적 안정성 검증이 필요하다.
- P99와 TPS는 pgbench 작업 부하에서 측정한 값이며, 실제 애플리케이션 커넥션 풀 대기와 타임아웃을 완전히 재현하지는 못한다.
- 시간대별 활성 커넥션, 디스크 입출력, CPU 사용률, 복제 지연을 함께 관찰하면 더 설득력 있는 비교가 가능하다.

## 6. 보고서 반영 방향

본 결과는 MigraGuard가 OSC를 대체한다는 근거가 아니라, 온라인 변경 작업을 시작하기 전 어느 시간대가 상대적으로 안전한지 판단하는 보조 도구로 활용될 수 있음을 보여주는 자료로 사용한다. 보고서에서는 다음과 같이 해석하는 것이 적절하다.

> PostgreSQL의 `CREATE INDEX CONCURRENTLY`는 쓰기 차단을 줄이는 온라인 변경 기능이지만, 높은 부하 조건에서는 P99 응답 지연이 크게 증가하였다. MigraGuard가 추천한 낮은 부하 시간대에서는 DDL 완료 시간이 가장 짧고 지연 증가도 낮은 절대값에 머물렀다. 다만 중간 부하 조건도 안정적으로 나타났으므로, 현재 forecast는 위험한 시간대를 피하는 데 의미가 있지만 최적 시간대를 정밀하게 식별하려면 추가 보정이 필요하다.
