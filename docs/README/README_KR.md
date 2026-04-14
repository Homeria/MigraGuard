# 🛡️ MigraGuard (PostgreSQL Migration Risk Gatekeeper)

> **"운영 트래픽을 모르는 DDL 배포는 시스템 마비의 시작입니다."**
>
> MigraGuard는 운영 데이터베이스의 실제 워크로드(TPS, 응답 시간, 커넥션 상태)와 마이그레이션 SQL을 교차 분석하여, **락(Lock) 경합으로 인한 서비스 장애를 배포 이전에 정량적으로 예측**하는 DevSecOps 도구입니다.

---

## 🏗️ 아키텍처 (Dual-Process Architecture)

MigraGuard v3.2는 데이터 수집의 지속성과 분석의 즉각성을 보장하기 위해 **이중 프로세스 구조**로 설계되었습니다.

1.  **MigraGuard Agent (Background Service)**
    *   운영 DB의 `pg_stat_statements` 및 시스템 뷰를 정기적으로 스캐닝.
    *   수집된 메트릭을 내장 SQLite(`migraguard.db`)에 시계열 데이터로 적재.
    *   최근 1시간 평균, 24시간 피크 트래픽 등 과거 이력 데이터 관리.
2.  **MigraGuard Analyze (CLI / CI-CD)**
    *   개발자의 SQL 파일을 AST(Abstract Syntax Tree)로 파싱하여 분석.
    *   Agent가 수집한 과거/실시간 지표를 기반으로 **5단계 리스크 엔진** 가동.
    *   위험 점수가 높을 경우 `Exit 1`을 반환하여 배포 파이프라인(GitHub Actions 등) 자동 차단.

---

## 🎯 캡스톤 디자인 구현 범위 (v3.2 핵심 기능)

본 프로젝트는 캡스톤 디자인 최종 결과물로서 아래 기능을 완벽히 구현하였습니다.

### 1. SQL 정적 분석 (AST Parsing)
*   `pganalyze/pg_query_go`를 활용하여 PostgreSQL 공식 파서와 100% 호환되는 구문 분석.
*   `ALTER TABLE`, `CREATE INDEX` 등 DDL 수행 시 **Table Rewrite(재작성)** 발생 여부 자동 판별.
*   대상 테이블 및 컬럼 존재 여부에 대한 스키마 사전 검증(Validation).

### 2. 5단계 정밀 리스크 모델 (Risk Engine)
단순한 룰 기반 탐지를 넘어, 대기 행렬 이론(Queuing Theory)을 응용한 수학적 모델링을 수행합니다.
*   **Step 1. $T_{ddl}$ 예측:** 테이블 크기 및 디스크 I/O 성능 기반 예상 작업 시간 산출.
*   **Step 2. $T_{block}$ 예측:** $T_{ddl}$ + P99 응답 시간 + 복제 지연(Lag)을 합산한 총 블로킹 시간 도출.
*   **Step 3. $C_{peak}$ 예측:** 블로킹 중 유입될 신규 커넥션 폭증량($\lambda \times T_{block}$) 계산.
*   **Step 4. $T_{rec}$ 예측:** 시스템 한계($C_{max}$) 초과 시 서비스 정상화까지 걸리는 회복 시간 예측.
*   **Step 5. Risk Score:** 최종 부하량 가중치($\lambda_{final}$)를 적용하여 **Safe / Warning / Danger** 판정.

### 3. 지능형 워크로드 분석
*   **Weighted TPS:** `Max(실시간, 1시간 평균 * 1.2, 24시간 피크 * 0.8)` 공식을 통한 보수적 위험 평가.
*   **Safe Window 추천:** 최근 24시간 트래픽 패턴을 분석하여 배포에 가장 안전한 시간대(저부하 시간) 자동 추천.

### 4. 유연한 리포팅 및 CI/CD 통합
*   **Console UI:** 터미널에서 즉시 확인 가능한 컬러풀한 테이블 리포트.
*   **Markdown Export:** PR 코멘트용 상세 분석 보고서 자동 생성.
*   **Custom Constants:** 인프라 사양(Disk I/O, Max Connections)에 맞춘 분석 상수 커스터마이징.

---

## 🐳 Docker로 실행하기 (Recommended)

MigraGuard는 운영 데이터베이스(PostgreSQL)와 함께 컨테이너 환경에서 실행하는 것이 가장 권장됩니다.

### 1. 전체 환경 실행 (Agent + Sample DB)
`docker-compose.yml`을 사용하여 샘플 데이터베이스와 분석 에이전트를 한 번에 띄웁니다.
```bash
docker compose up -d
```

### 2. 특정 SQL 분석 실행 (Analyze)
에이전트가 실행 중인 상태에서, 공유 볼륨을 통해 실시간 데이터를 기반으로 분석을 수행합니다.
```bash
# 로컬의 SQL 파일을 컨테이너를 통해 분석
docker compose run --rm analyze-shell analyze /app/code/migrations/001_heavy_alter.sql
```

---

## 🛠️ 시작하기 (Quick Start)

### 1. 전제 조건
*   **PostgreSQL:** `pg_stat_statements` 확장 설치 및 활성화 필요.
*   **Go:** v1.25 이상 권장.

### 2. 설정 파일 작성 (`migraguard.yaml`)
프로젝트 루트 또는 실행 경로에 설정 파일을 작성합니다.

```yaml
database:
  postgres: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
  sqlite: "./migraguard.db"

agent:
  interval: "1m"       # 지표 수집 주기
  retention_days: 7    # 데이터 보관 기간

risk:
  disk_io: 104857600   # 100MB/s (Table Rewrite 시간 계산용)
  c_max: 500           # DB 최대 커넥션 수
  mu_max: 5000.0       # 시스템 한계 TPS
  t_meta: 100.0        # 메타데이터 변경 기본 지연시간 (ms)
```

### 3. 에이전트 실행 (수집 모드)
운영 환경 또는 모니터링 서버에서 상시 실행합니다.
```bash
./migraguard agent
```

### 4. 리스크 분석 실행 (분석 모드)
마이그레이션 SQL 파일을 대상으로 분석을 수행합니다.
```bash
# 기본 콘솔 출력
./migraguard analyze ./migrations/001_heavy_alter.sql

# 마크다운 파일로 저장 (CI용)
./migraguard analyze ./migrations/001_heavy_alter.sql --format markdown > report.md

# 상세 로그 포함
./migraguard analyze ./migrations/001_heavy_alter.sql --verbose
```

---

## 🚀 CI/CD 적용 예시 (GitHub Actions)

```yaml
steps:
  - name: Run MigraGuard Analysis
    run: |
      ./migraguard analyze ./deploy/schema_update.sql --format markdown > risk_report.md
    continue-on-error: false # 위험(Danger) 판정 시 빌드 중단

  - name: Comment PR
    uses: thollander/actions-comment-pull-request@v2
    with:
      filePath: risk_report.md
```

---

## 🧪 테스트 환경 구축 (Testing)

MigraGuard의 리스크 분석을 정확히 테스트하기 위해서는 `pg_stat_statements`가 활성화된 PostgreSQL이 필요합니다.

### 1. PostgreSQL 설정
`postgresql.conf` 파일에 아래 설정을 추가하거나, Docker 실행 시 옵션을 부여합니다.
```bash
# Docker 실행 예시
docker run -d --name mg-db -e POSTGRES_PASSWORD=pass -p 5432:5432 postgres:15-alpine -c shared_preload_libraries=pg_stat_statements
```

접속 후 확장을 생성합니다.
```sql
CREATE EXTENSION pg_stat_statements;
```

### 2. 시나리오 테스트 케이스
`migrations/` 폴더 내의 테스트 케이스를 사용하여 분석 엔진의 반응을 확인하세요.

| 리스크 | SQL 파일 | 설명 |
| :--- | :--- | :--- |
| **Safe** | `001_safe_set_default.sql` | 단순 기본값 설정 (Metadata only) |
| **Warning** | `002_warning_add_index.sql` | 인덱스 생성 (Lock competition) |
| **Danger** | `003_danger_rewrite_type.sql` | 컬럼 타입 변경 (Table Rewrite) |

---

## 📜 라이선스 및 제작
*   **제작:** Homeria / MigraGuard Team
*   **기술 스택:** Golang, PostgreSQL, SQLite, Cobra, Viper, pg_query_go
