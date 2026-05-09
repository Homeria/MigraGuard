# 📜 MigraGuard 스크립트 상세 매뉴얼

이 문서는 MigraGuard의 자동화 스크립트 체계와 각 스크립트의 상세 사용법을 KR/EN 이국어로 제공합니다.

---

## 📂 1. Sandbox 그룹 (데이터 생성)
위치: `scripts/{platform}/sandbox/`

### 1.1. gen-db-from-scenario
- **역할**: 단일 YAML 시나리오로 SQLite 샌드박스 생성.
- **PowerShell**: `.\scripts\ps1\sandbox\gen-db-from-scenario.ps1 -Scenario 01_steady.yaml`

### 1.2. gen-db-from-all-scenarios
- **역할**: 모든 시나리오 일괄 생성.
- **Bash**: `./scripts/sh/sandbox/gen-db-from-all-scenarios.sh`

---

## 🔬 2. Analyze 그룹 (리스크 분석)
위치: `scripts/{platform}/analyze/`

### 2.1. run-analysis-from-sandbox
- **역할**: 생성된 샌드박스 파일을 소스로 오프라인 분석 수행.
- **CMD**: `scripts\cmd\analyze\run-analysis-from-sandbox.bat steady.db add_col.sql`

### 2.2. run-batch-research-report
- **역할**: 전수 교차 분석 및 연구 리포트 생성.
- **PowerShell**: `.\scripts\ps1\analyze\run-batch-research-report.ps1`

---

## 📊 3. 시각화 도구
- **시나리오 지표 시각화**: `python tools/visualization/sandbox/visualize_all_scenarios.py`
- **분석 결과 시각화**: `python tools/visualization/analyze/visualize_all_reports.py`
