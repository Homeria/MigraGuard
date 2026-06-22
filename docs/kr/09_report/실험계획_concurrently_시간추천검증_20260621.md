# CREATE INDEX CONCURRENTLY 기반 시간대 추천 검증 실험 계획

## 목적

본 실험의 목적은 MigraGuard가 forecast로 추천한 시간대가 실제 온라인 변경 작업을 수행하기에 적합한 시간대인지 확인하는 것이다. PostgreSQL 환경에서는 MySQL 기반 OSC 도구인 `gh-ost`, `pt-online-schema-change`를 직접 적용하기 어렵기 때문에, PostgreSQL이 공식 제공하는 DBMS 레벨 온라인 변경 기능인 `CREATE INDEX CONCURRENTLY`를 사용한다.

핵심 질문은 다음과 같다.

> MigraGuard가 추천한 낮은 위험 시간대에서 `CREATE INDEX CONCURRENTLY`를 실행했을 때, 다른 시간대보다 P99 응답 지연 증가, TPS 하락, 실패 트랜잭션 및 DDL 완료 시간이 더 작게 나타나는가?

## 실험 대상

- DDL: `experiments/ddl/026_safe_concurrent_orders_composite_index.sql`
- 작업: `orders(status, created_at)` 복합 인덱스를 `CREATE INDEX CONCURRENTLY`로 생성
- 대상 테이블: `orders`
- 작업 부하: `orders_mixed`
- DB 규모: 기본 fixture + 추가 주문 데이터

## 비교 조건

MigraGuard forecast 결과에서 DDL-2는 02시를 Safe 추천 시간대로 제시하였다. 이를 기준으로 다음 세 부하 조건을 비교한다.

| 조건 | 의미 | Target TPS |
|---|---|---:|
| 추천 시간대 재현 | MigraGuard가 추천한 저부하 시간대 | 10 |
| 중간 부하 시간대 재현 | 일반 업무 시간대 수준 | 1000 |
| 고부하 시간대 재현 | MigraGuard가 Danger로 볼 가능성이 큰 시간대 | 5000 |

## 측정 지표

- DDL 완료 시간
- 실제 TPS
- P50, P95, P99 응답 지연
- 실패 트랜잭션 수
- 기준선 대비 P99 증가율
- 기준선 대비 TPS 변화

## 해석 기준

MigraGuard 추천 시간대에서 DDL 완료 시간이 짧고 P99 증가율 및 TPS 하락이 작게 나타나면, forecast 결과가 온라인 변경 작업의 시작 시점 선택에 활용될 수 있음을 의미한다. 반대로 추천 시간대와 실측 최적 시간대가 다르면, RiskScore 계산식과 시간대별 지표 보정이 필요함을 의미한다.

## 한계

본 실험은 완전한 OSC 도구를 직접 비교한 것이 아니다. `CREATE INDEX CONCURRENTLY`는 PostgreSQL이 제공하는 DBMS 레벨 온라인 변경 기능이며, gh-ost나 pt-online-schema-change처럼 shadow table 복사, 변경분 동기화, cut-over, throttle/pause 횟수를 모두 포함하지 않는다. 따라서 본 실험은 "온라인 변경 작업의 시작 시간대 추천"을 예비적으로 검토하는 실험으로 해석한다.
