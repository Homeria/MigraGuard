# MigraGuard 기술 사양 및 아키텍처 (상세)

## 1. 시스템 동작 흐름 (Core Engine Flow)
1. **백그라운드 수집 (Background Collection):**
   - 고루틴 기반 데몬이 `robfig/cron` 또는 Ticker를 사용하여 주기적으로 운영 DB의 `pg_stat_statements` 스냅샷을 캡처합니다.
   - 캡처된 데이터는 `internal/db/workload.go`를 통해 로컬 SQLite 저장소에 누적됩니다.
2. **분석 단계 (Analyze Phase):**
   - 사용자가 `analyze` 명령으로 마이그레이션 SQL을 입력합니다.
   - `internal/parser/ast.go`가 SQL을 AST로 변환하여 타겟 테이블을 식별하고, 해당 테이블에 필요한 Lock Level을 판별합니다.
3. **위험도 평가 (Risk Evaluation):**
   - SQLite에 저장된 시계열 트래픽 데이터를 분석하여, 해당 테이블의 평소 트래픽 밀도를 파악합니다.
   - 중앙값(Median) 연산을 통해 '최적의 안전 배포 시간대(Safe Window)'를 추천합니다.
   - 현재 트래픽이 임계치를 초과하거나 락 레벨이 높은 경우 `Risk Score`를 산출합니다.
4. **결과 리포팅 및 게이트키핑:**
   - CLI 결과 출력 및 GitHub PR 봇 연동을 통한 마크다운 코멘트 작성.
   - 위험 수준이 'Danger'인 경우 프로세스를 중단시켜 배포를 차단합니다.

## 2. 주요 모듈 및 기술 상세
- **데이터 저장소 (Storage):** 외부 의존성을 줄이기 위해 로컬 SQLite(`mattn/go-sqlite3`)를 채택하여 WAL 모드로 시계열 데이터를 관리합니다.
- **파서 (Parser):** PostgreSQL 정식 C 파서를 포팅한 `pg_query_go`를 사용하여 최신 PostgreSQL 문법을 완벽히 지원합니다.
- **동시성 모델:** Go의 `channel`과 `goroutine`을 활용하여 메인 프로세스의 블로킹 없이 백그라운드 수집을 수행합니다.

## 3. MigraGuard v3.0 정적/동적 통합 위험도 산출 수학적 모델

본 섹션은 DDL 실행 시 발생할 수 있는 대기행렬(Queue)의 급격한 증가와 이로 인한 연쇄 장애(Cascading Failure)를 예측하기 위한 정량적 분석 모델을 기술합니다.

### 3.1. 위험도 산출 공식 (5단계)

#### Step 1. DDL 물리적 소요 시간 추정 ($T_{ddl}$)
$$T_{ddl} = \left( F_{rewrite} \times \frac{S_{table}}{Disk_{IO}} \right) + \left( (1 - F_{rewrite}) \times T_{meta} \right)$$
*   **설명**: AST 파싱 결과를 바탕으로 테이블 재기록(Table Rewrite) 발생 여부를 판단합니다. 발생 시 물리적 디스크 I/O 처리 시간을 계산하며, 단순 메타데이터 변경 시에는 고정된 상수 시간을 부여합니다.

#### Step 2. 총 블로킹 시간 산출 ($T_{block}$)
$$T_{block} = T_{p99} + T_{ddl} + Lag_{repl}$$
*   **설명**: 락 경합으로 인해 후행 쿼리가 차단되는 총 시간을 의미합니다. 상위 1%($P99$) 느린 쿼리의 대기 시간, DDL 실행 시간, 그리고 복제 지연(Replication Lag) 시간을 합산하여 산출합니다.

#### Step 3. 락 해제 직후 큐 스파이크량 산출 ($C_{peak}$)
$$C_{peak} = C_{active} + (\lambda \times T_{block})$$
*   **설명**: 락이 유지되는($T_{block}$) 동안 처리되지 못하고 대기열에 누적된 최대 커넥션 요구량입니다.

#### Step 4. 시스템 회복 소요 시간 및 타임아웃 지연 시간 산출 ($T_{rec}, T_{delay\_max}$)
$$T_{rec} = \max\left( 0, \frac{C_{peak} - C_{max}}{\mu_{max} - \lambda} \right)$$
$$T_{delay\_max} = T_{block} + T_{rec}$$
*   **설명**: 축적된 큐 스파이크($C_{peak}$)를 DB의 최대 처리량($\mu_{max}$)으로 해소하는 데 걸리는 시간입니다. 만약 유입량($\lambda$)이 최대 처리량보다 클 경우($\lambda \geq \mu_{max}$), 시스템은 영구적 장애 상태로 간주됩니다.

#### Step 5. 최종 커넥션 고갈 위험도 ($RiskScore$)
$$RiskScore(\%) = \left( \frac{C_{peak}}{C_{max}} \right) \times 100$$
*   **설명**: 시스템이 수용 가능한 최대 커넥션($C_{max}$) 대비 예측된 큐 스파이크량의 비율을 통해 서비스 불능(OOM, Connection Exhaustion) 가능성을 점수화합니다.

### 3.2. 시스템 변수 정의 (Variable Definitions)

| 변수 | 명칭 | 설명 | 수집 출처/분류 |
| :--- | :--- | :--- | :--- |
| $L_{type}$ | 락 강도 계수 | DDL이 요구하는 락의 종류 (AccessExclusiveLock 등) | `pg_query_go` (정적 파싱) |
| $F_{rewrite}$ | 테이블 재기록 플래그 | DDL 실행 시 테이블 풀 스캔 및 복사 발생 여부 (1 또는 0) | `pg_query_go` (정적 파싱) |
| $S_{table}$ | 타겟 테이블 물리 크기 | 타겟 테이블의 현재 디스크 점유 용량 (Bytes) | `pg_relation_size()` (동적 DB 상태) |
| $Disk_{IO}$ | DB 디스크 I/O 성능 | 스토리지의 초당 처리 속도 (Bytes/sec) | 인프라 설정값 (상수) |
| $T_{meta}$ | 메타데이터 처리 시간 | 단순 카탈로그 업데이트에 소요되는 평균 시간 | 실측 경험치 (상수) |
| $T_{p99}$ | P99 쿼리 실행 시간 | 타겟 테이블에 실행 중인 상위 1%의 꼬리 지연 시간 | `pg_stat_statements` (동적 트래픽) |
| $Lag_{repl}$ | 복제 지연 시간 | Master-Replica 간 WAL 동기화 지연 시간 | `pg_stat_replication` (동적 DB 상태) |
| $\lambda$ | 쿼리 유입량 (TPS) | 초당 타겟 테이블에 들어오는 평균 쿼리 수 | `pg_stat_statements` (동적 트래픽) |
| $\mu_{max}$ | 최대 초당 처리량 | DB가 에러 없이 처리 가능한 초당 최대 쿼리 수 | 인프라 설정 및 부하테스트 (상수) |
| $C_{max}$ | 최대 커넥션 풀 크기 | PgBouncer 등에 설정된 Max Connection 제한치 | 환경 변수 (상수) |
| $C_{active}$ | 현재 활성 커넥션 수 | 타겟 테이블을 점유/대기 중인 트랜잭션 수 | `pg_stat_activity` (동적 트래픽) |
| $T_{timeout}$ | API 타임아웃 임계치 | 백엔드 서버에 설정된 API 강제 종료 시간 | 애플리케이션 환경 변수 (상수) |
