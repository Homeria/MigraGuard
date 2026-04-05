# 🛡️ MigraGuard v3.2 초정밀 로직 가이드 (Caller-Callee Deep Dive)

본 문서는 MigraGuard의 실행부터 종료까지, 모든 함수 호출과 데이터 반환 과정을 소스 코드 레벨에서 기술합니다.

---

## 1. 뼈대 및 부트스트랩 (The Skeleton)

### [1-1] 진입점: `main.go`
- **Caller**: OS / User Shell
- **Action**: `cmd.Execute()` 호출.
- **Callee**: `cmd/root.go` -> `rootCmd.Execute()`

### [1-2] 설정 로드: `cmd/root.go`
- **Function**: `initConfig()`
- **Action**: `viper`를 통해 `migraguard.yaml` 읽기 및 `GlobalConfig` 객체 채우기.
- **Return**: 전역 설정 완료. 이후 모든 레이어에서 `GlobalConfig` 참조 가능.

---

## 2. Agent 모드: 상시 지표 수집 (The Heartbeat)

### [2-1] 커맨드 핸들러: `cmd/migraguard/agent.go`
- **Function**: `agentCmd.Run`
- **Calls**:
    1.  `db.NewPostgresAdapter(url)` (`internal/db/connection.go`): PG 연결 풀 생성.
    2.  `db.NewSQLiteAdapter(path)` (`internal/db/activity.go`): SQLite 파일 오픈 및 `initSchema()`로 테이블 생성.
    3.  `service.NewAgentService(pg, sqlite, ...)` (`internal/service/agent.go`): 서비스 객체 생성 (DI 주입).
    4.  **`svc.Run(ctx)`** 호출.

### [2-2] 도메인 서비스: `internal/service/agent.go`
- **Function**: `AgentService.Run(ctx)`
- **Calls**:
    1.  `db.NewCollector(pg, sqlite, interval)` (`internal/db/collector.go`): 컬렉터 객체 생성.
    2.  `collector.SetRetentionDays(days)`: 데이터 보존 기간 설정.
    3.  **`collector.Start(ctx)`**: 핵심 수집 루프 시작.

### [2-3] 백그라운드 컬렉터: `internal/db/collector.go`
- **Function**: `collector.collect(ctx)` (매 `interval`마다 실행)
- **Flow**:
    1.  **`pg.FetchWorkload(ctx)`** (`internal/db/workload.go`):
        - `pg_stat_statements` 조회 후 `[]WorkloadSnapshot` 반환.
    2.  **`sqlite.SaveSnapshots(snapshots)`** (`internal/db/activity.go`):
        - 트랜잭션 시작 → `INSERT INTO workload_snapshots` 실행 → `Commit`.
    3.  **Loop: `pg.GetTableDynamicMetrics(ctx, table)`**:
        - `pg_total_relation_size()`, `pg_stat_replication`, `pg_stat_activity` 조회.
        - `*TableDynamicMetrics` 객체 생성 및 반환.
    4.  **`sqlite.SaveTableMetrics(metrics)`**:
        - `INSERT INTO table_metrics` 실행.
    5.  **`sqlite.PurgeOldSnapshots(days)`**:
        - `DELETE` 쿼리로 노후 데이터 삭제 후 **`VACUUM`** 호출로 물리 공간 회수.

---

## 3. Analyze 모드: 즉각 리스크 분석 (The Brain)

### [3-1] 커맨드 핸들러: `cmd/migraguard/analyze.go`
- **Function**: `analyzeCmd.Run`
- **Action**: 인프라(Adapters) 초기화 후 **`svc.Run(ctx, task)`** 호출.

### [3-2] 도메인 서비스: `internal/service/analyze.go`
- **Function**: `AnalyzeService.Run(ctx, task)`
- **Flow**:
    1.  **`os.ReadFile(task.SQLPath)`**: SQL 파일 읽기 (`string` 반환).
    2.  **`parser.ParseSQL(sql)`** (`internal/parser/ast.go`):
        - `pg_query.Parse(sql)` 호출 → AST 생성.
        - `handleNode()` 순회하며 `[]AnalysisResult` 반환 (테이블명, 락 레벨, `RewriteRequired` 포함).
    3.  **Loop: `pg.ValidateSchema(res.TableName, res.Columns)`** (`internal/db/workload.go`):
        - `information_schema` 조회하여 테이블/컬럼 존재 여부 확인 (`bool` 반환).
    4.  **`riskEngine.AnalyzeRisk(ctx, res)`** (`internal/engine/risk.go`): 분석 실행.

### [3-3] 리스크 엔진: `internal/engine/risk.go`
- **Function**: `AnalyzeRisk(ctx, analysis)`
- **Detailed Flow**:
    1.  **`pg.GetTableDynamicMetrics()`**: Postgres에서 실시간 기초 지표(Size, Lag 등) 로드.
    2.  **`sqlite.GetRecentTPSDelta(table)`**: SQLite에서 최근 1분간의 델타 TPS($\lambda_{curr}$) 계산.
    3.  **`sqlite.GetTableBaselineStats(table)`**: SQLite에서 최근 24시간 피크($\lambda_{peak}$) 및 1시간 평균($\lambda_{avg}$) 로드.
    4.  **TPS 가중치 결정**: $\lambda_{final} = \max(\lambda_{curr}, \lambda_{avg} \times 1.2, \lambda_{peak} \times 0.8)$
    5.  **수식 계산**:
        - `T_ddl`: `RewriteRequired`이면 `(Size / DiskIO) * 1000` 아니면 `TMeta`.
        - `T_block`: `P99 + T_ddl + ReplicationLag`.
        - `C_peak`: `ActiveConns + (Lambda_final/1000 * T_block)`.
        - `RiskScore`: `(C_peak / CMax) * 100`.
    6.  **`RiskAnalysisReport` 생성 및 반환**.

### [3-4] 리포팅: `internal/reporter/`
- **Function**: `rpt.Write(results, reports)`
- **Callee**: `ConsoleReporter.Write()` 또는 `MarkdownReporter.Write()`
- **Action**: 가공된 데이터를 `fmt.Printf` 또는 `os.WriteFile`로 출력.

---

## 4. 레이어별 데이터 객체 (Data Objects)

| 레이어 | 객체명 | 주요 역할 |
| :--- | :--- | :--- |
| **Parser** | `AnalysisResult` | SQL 정적 분석 결과 (테이블, 락, 재기록 여부) |
| **Adapters** | `WorkloadSnapshot` | `pg_stat_statements` 스냅샷 데이터 |
| **Adapters** | `TableDynamicMetrics` | DB 실시간 지표 (Size, Lag, Conns, TPS) |
| **Engine** | `RiskAnalysisReport` | 최종 분석 결과 (Score, Time, Level) |
| **Service** | `AnalysisResponse` | CLI로 전달되는 최종 결과 묶음 |

---

## 5. 요약: 데이터의 여정 (Data Journey)
1.  **Agent**가 **Postgres**에서 데이터를 퍼 올려 **SQLite**에 담아둔다.
2.  **Analyze**가 **SQL 파일**을 읽어 "무엇을 할지" 파악한다.
3.  **Analyze**가 **SQLite**에서 저장된 과거 트래픽을 가져와 "지금 하면 위험한지" 계산한다.
4.  결과가 **Danger**면 `os.Exit(1)`로 프로세스를 강제 종료하여 배포를 차단한다.
