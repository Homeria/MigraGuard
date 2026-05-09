# 📊 MigraGuard 연구 데이터 추출 및 분석 가이드

본 문서는 MigraGuard의 **일괄 배치 분석(Batch Research)** 기능을 통해 대규모 실험 데이터를 CSV로 추출하고, 시각화 도구를 활용해 학술적 통계 자료를 생성하는 방법을 안내합니다.

---

## 1. 개요 (Overview)

MigraGuard는 단순한 리스크 탐지 도구를 넘어, DDL과 트래픽 간의 상관관계를 정량적으로 연구할 수 있는 실험 플랫폼을 제공합니다. 

*   **전수 조사(Exhaustive Analysis)**: 모든 시나리오와 테스트 DDL을 교차 분석하여 수백 개의 정밀 데이터를 일괄 확보.
*   **시각화 자동화**: Python 스크립트를 통한 부하 프로필 및 리스크 히트맵 자동 생성.

---

## 2. 데이터 분석 프로세스 (Workflow)

### Step 1: 연구용 데이터 셋 구축
모든 시나리오에 대해 SQLite 샌드박스 파일을 생성합니다.
```powershell
.\scripts\ps1\sandbox\gen-db-from-all-scenarios.ps1
```

### Step 2: 전수 조사 및 CSV 추출
구축된 데이터 셋과 `experiments/ddl` 폴더의 모든 케이스를 결합하여 분석을 실행합니다.
```powershell
.\scripts\ps1\analyze\run-batch-research-report.ps1
```
*   **결과**: `experiments\reports\research_results.csv` 파일 생성.

### Step 3: 데이터 시각화 (Visualization)
제공되는 Python 스크립트를 사용하여 그래프를 생성합니다.

**A. 리스크 히트맵 생성**
```bash
python tools/visualization/analyze/visualize_all_reports.py
```

**B. 부하 프로필 시각화**
```bash
python tools/visualization/sandbox/visualize_all_scenarios.py
```

---

## 3. 연구 활용 팁 (Research Tips)

### 3.1. 엑셀 피벗 테이블 활용
추출된 CSV의 `RiskScore`를 기준으로 시나리오별 위험도 추이를 분석할 수 있습니다. 특히 `TableSize`와 `T_ddl` 간의 상관관계를 차트화하여 시스템의 예측 정확도를 증명할 수 있습니다.

### 3.2. 정밀 분석 예시 (Python)
연구원들은 Pandas를 사용하여 특정 위험군(Danger) 사례들만 필터링하여 인프라 임계치($C_{max}$)와의 관계를 정량적으로 분석할 수 있습니다.
