# 📜 MigraGuard 스크립트 상세 매뉴얼 (Script Manual)

이 문서는 MigraGuard의 자동화 스크립트 체계와 각 스크립트의 상세 사용법, 인자(Parameter), 그리고 플랫폼별 실행 예시를 다룹니다.

---

## 📂 1. Sandbox 그룹 (데이터 생성 및 시뮬레이션)
위치: `scripts/{platform}/sandbox/`

### 1.1. gen-db-from-scenario
*   **역할**: 단일 YAML 시나리오를 실행하여 SQLite 샌드박스 DB(`.db`)를 생성합니다.
*   **인자**: `Scenario` (파일명), `Force` (강제 재생성 여부)
*   **실행 예시**:
    *   **PowerShell**: `.\scripts\ps1\sandbox\gen-db-from-scenario.ps1 -Scenario 01_steady.yaml -Force`
    *   **CMD**: `scripts\cmd\sandbox\gen-db-from-scenario.bat 01_steady.yaml --force`
    *   **Bash**: `./scripts/sh/sandbox/gen-db-from-scenario.sh 01_steady.yaml --force`

### 1.2. gen-db-from-all-scenarios
*   **역할**: 모든 시나리오에 대해 DB를 일괄 생성합니다.
*   **실행 예시**:
    *   **PowerShell**: `.\scripts\ps1\sandbox\gen-db-from-all-scenarios.ps1 -Force`
    *   **CMD**: `scripts\cmd\sandbox\gen-db-from-all-scenarios.bat --force`
    *   **Bash**: `./scripts/sh/sandbox/gen-db-from-all-scenarios.sh --force`

### 1.3. gen-csv-from-scenario
*   **역할**: DB 미생성, 즉시 시뮬레이션 지표 CSV 추출.
*   **인자**: `Scenario` (파일명), `Output` (저장 경로)
*   **실행 예시**:
    *   **PowerShell**: `.\scripts\ps1\sandbox\gen-csv-from-scenario.ps1 -Scenario 03_spike.yaml -Output spike_metrics.csv`
    *   **CMD**: `scripts\cmd\sandbox\gen-csv-from-scenario.bat 03_spike.yaml spike_metrics.csv`
    *   **Bash**: `./scripts/sh/sandbox/gen-csv-from-scenario.sh 03_spike.yaml spike_metrics.csv`

### 1.4. gen-csv-from-all-scenarios
*   **역할**: 모든 시나리오의 지표를 `experiments/reports/metrics/`에 일괄 추출.
*   **실행 예시**:
    *   **PowerShell**: `.\scripts\ps1\sandbox\gen-csv-from-all-scenarios.ps1`
    *   **CMD**: `scripts\cmd\sandbox\gen-csv-from-all-scenarios.bat`
    *   **Bash**: `./scripts/sh/sandbox/gen-csv-from-all-scenarios.sh`

---

## 🔬 2. Analyze 그룹 (리스크 분석 및 리포팅)
위치: `scripts/{platform}/analyze/`

### 2.1. run-analysis-from-sandbox (오프라인 분석)
*   **역할**: 생성된 `.db` 파일을 소스로 SQL 위험도 분석.
*   **인자**: `DB` (파일명), `SQL` (파일명), `Output` (형식), `ReportPath` (저장경로 - PS1 전용)
*   **실행 예시**:
    *   **PowerShell**: `.\scripts\ps1\analyze\run-analysis-from-sandbox.ps1 -DB steady.db -SQL add_col.sql -Output csv -ReportPath result.csv`
    *   **CMD**: `scripts\cmd\analyze\run-analysis-from-sandbox.bat steady.db add_col.sql csv > result.csv`
    *   **Bash**: `./scripts/sh/analyze/run-analysis-from-sandbox.sh steady.db add_col.sql csv > result.csv`

### 2.2. run-analysis-from-live (온라인 분석)
*   **역할**: 실운영 Postgres 접속을 통한 실시간 분석.
*   **인자**: `SQL` (파일명), `DBUrl`, `SQLite`, `Output`
*   **실행 예시**:
    *   **PowerShell**: `.\scripts\ps1\analyze\run-analysis-from-live.ps1 -SQL patch.sql -DBUrl "postgres://..."`
    *   **CMD**: `scripts\cmd\analyze\run-analysis-from-live.bat patch.sql "postgres://..."`
    *   **Bash**: `./scripts/sh/analyze/run-analysis-from-live.sh patch.sql "postgres://..."`

### 2.3. run-batch-research-report (전수 리서치)
*   **역할**: 모든 DB x 모든 SQL 교차 배치 분석.
*   **실행 예시**:
    *   **PowerShell**: `.\scripts\ps1\analyze\run-batch-research-report.ps1`
    *   **CMD**: `scripts\cmd\analyze\run-batch-research-report.bat`
    *   **Bash**: `./scripts/sh/analyze/run-batch-research-report.sh`

---

## 💡 플랫폼별 인자 전달 방식

*   **PowerShell**: `-Name Value` 형태의 명명된 파라미터를 지원하여 순서에 상관없이 입력 가능합니다.
*   **CMD/Bash**: 미리 정의된 순서(Positional Argument)대로 입력해야 합니다. (예: DB명 다음에 SQL명)

---

## 📊 3. 시각화 도구 (Python Visualization)
위치: `tools/visualization/`
목적: 추출된 CSV 데이터를 바탕으로 분석용 그래프 및 히트맵을 생성합니다.

### 3.1. 환경 준비 (Prerequisites)
```bash
pip install pandas matplotlib seaborn
```

### 3.2. Sandbox 시각화 (시나리오 부하 지표)
시뮬레이션에서 생성된 물리적 지표(TPS, I/O 등)를 시각화합니다. 결과는 `[파일명]_plots/` 폴더에 7개의 개별 이미지로 저장됩니다.

*   **[1개 분석]** `plot_scenario_metrics.py`
    ```bash
    python tools/visualization/sandbox/plot_scenario_metrics.py experiments/reports/metrics/01_steady_metrics.csv
    ```
*   **[일괄 분석]** `visualize_all_scenarios.py`
    ```bash
    python tools/visualization/sandbox/visualize_all_scenarios.py
    ```

### 3.3. Analyze 시각화 (리스크 엔진 결과)
리스크 엔진이 계산한 통계 데이터(점수, 등급 등)를 시각화합니다. 결과는 `[파일명]_analysis.png` 한 장의 대시보드 이미지로 저장됩니다.

*   **[1개 분석]** `plot_research_report.py`
    ```bash
    python tools/visualization/analyze/plot_research_report.py experiments/reports/research_results.csv
    ```
*   **[일괄 분석]** `visualize_all_reports.py`
    ```bash
    python tools/visualization/analyze/visualize_all_reports.py
    ```

---
**Tip**: 모든 상태 로그는 `os.Stderr`로 출력되므로, 리다이렉션(`>`) 시 데이터만 깨끗하게 추출됩니다.
