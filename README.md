# 🛡️ MigraGuard (PostgreSQL Migration Risk Gatekeeper)

> **"운영 트래픽을 모르는 DDL 배포는 시스템 마비의 시작입니다."**
>
> MigraGuard는 운영 데이터베이스의 실제 워크로드(TPS, 응답 시간, 커넥션 상태)와 마이그레이션 SQL을 교차 분석하여, **락(Lock) 경합으로 인한 서비스 장애를 배포 이전에 정량적으로 예측**하는 DevSecOps 도구입니다.

---

## 🏗️ 아키텍처 (Dual-Process Architecture)

MigraGuard v3.8은 데이터 수집의 지속성과 분석의 즉각성을 보장하기 위해 **이중 프로세스 구조**로 설계되었습니다.

1.  **MigraGuard Agent (Background Service)**
    *   운영 DB의 `pg_stat_statements` 및 시스템 뷰를 정기적으로 스캐닝.
    *   수집된 메트릭을 내장 SQLite(`migraguard.db`)에 시계열 데이터로 적재.
2.  **MigraGuard Analyze (CLI / CI-CD)**
    *   개발자의 SQL 파일을 AST(Abstract Syntax Tree)로 파싱하여 분석.
    *   Agent가 수집한 지표 또는 **시뮬레이션 샌드박스 데이터**를 기반으로 분석 수행.
    *   위험 점수가 높을 경우 `Exit 1`을 반환하여 배포 파이프라인 자동 차단.

---

## 🎯 핵심 기능 (v3.8)

### 1. 5단계 정밀 리스크 모델 (Risk Engine)
*   **Step 1. $T_{ddl}$ 예측:** 테이블 크기 및 디스크 I/O 성능 기반 예상 작업 시간 산출.
*   **Step 2. $T_{block}$ 예측:** $T_{ddl}$ + P99 응답 시간 + 복제 지연(Lag) 합산.
*   **Step 3. $C_{peak}$ 예측:** 블로킹 중 유입될 신규 커넥션 폭증량 계산.
*   **Step 4. $T_{rec}$ 예측:** 시스템 한계 초과 시 서비스 정상화까지의 회복 시간 예측.
*   **Step 5. Risk Score:** 보수적 가중치를 적용하여 **Safe / Warning / Danger** 판정.

### 2. 시뮬레이션 샌드박스 및 가상화 (Research Ready)
*   **Scenario Seeder:** YAML 기반으로 7일치 가상 데이터를 1초 만에 생성.
*   **Offline Analysis:** 실제 DB 없이 샌드박스 데이터만으로 오프라인 리스크 평가 수행 (`--sandbox`).
*   **Multi-Platform Automation:** Windows CMD, PowerShell, Bash용 통합 자동화 스크립트 제공.

### 3. 지능형 워크로드 분석
*   **Weighted TPS:** 실시간, 1시간 평균, 24시간 피크 상황을 교차 분석하여 최악의 시나리오 가정.
*   **Safe Window 추천:** 배포에 가장 안전한 저부하 시간대 자동 추천.

---

## 🧪 독립형 시나리오 실험 (Sandbox Mode)

실제 DB 없이 가상의 장애 상황을 재현하고 검증할 때 사용합니다.

```cmd
:: 1. 샌드박스 데이터 생성 및 분석 (윈도우 CMD)
scripts\cmd\sandbox.bat 03_spike_flash_sale.yaml 011_danger_rewrite_order_no.sql

:: 2. 모든 연구 시나리오 일괄 시드 주입
scripts\cmd\seed_all.bat
```

모든 실험 자산은 `experiments/` 워크스페이스(Scenarios, DDL, Data)에서 통합 관리됩니다.

---

## 🛠️ 시작하기 (Quick Start)

### 1. 에이전트 실행 (수집 모드)
```bash
./migraguard agent
```

### 2. 리스크 분석 실행 (분석 모드)
```bash
# 실시간 데이터 기반 분석
./migraguard analyze ./migrations/001_heavy_alter.sql

# 시뮬레이션 샌드박스 기반 오프라인 분석
./migraguard analyze ./experiments/ddl/001_sql.sql --sandbox ./experiments/data/spike.db
```

---

## 📚 상세 문서 (Documentation)

- **[시뮬레이션 샌드박스 운용 가이드]** [`docs/04_guides/sandbox_manual.md`](docs/04_guides/sandbox_manual.md)
- **[시스템 아키텍처 개요]** [`docs/02_architecture/system_overview.md`](docs/02_architecture/system_overview.md)
- **[구현 상세 명세]** [`docs/03_implementation/parser_logic.md`](docs/03_implementation/parser_logic.md)

---

## 📜 라이선스 및 제작
*   **제작:** Homeria / MigraGuard Team
*   **기술 스택:** Golang, PostgreSQL, SQLite, Cobra, Viper, pg_query_go
