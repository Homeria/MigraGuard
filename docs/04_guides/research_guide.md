# 📊 MigraGuard 연구 데이터 추출 및 분석 가이드 (Research Guide)

본 문서는 MigraGuard의 **일괄 배치 분석(Batch Research)** 기능을 통해 대규모 실험 데이터를 CSV로 추출하고, 제공되는 시각화 도구를 활용해 논문이나 보고서용 통계 자료를 생성하는 방법을 안내합니다.

---

## 1. 개요 (Overview)

MigraGuard는 단순한 리스크 탐지 도구를 넘어, DDL과 트래픽 간의 상관관계를 정량적으로 연구할 수 있는 실험 플랫폼을 제공합니다. 

*   **전수 조사(Exhaustive Analysis)**: 모든 시뮬레이션 시나리오에 대해 모든 테스트 DDL을 교차 분석하여 수백 개의 정밀 데이터를 한 번에 확보합니다.
*   **시각화 자동화**: 추출된 CSV 데이터를 바탕으로 부하 프로필 및 리스크 히트맵을 생성하는 Python 스크립트를 제공합니다.

---

## 2. 데이터 확보 및 분석 프로세스 (Workflow)

가장 권장되는 **PowerShell(ps1)** 환경을 기준으로 설명합니다. (CMD 및 Bash 스크립트도 구조는 동일합니다.)

### Step 1: 연구용 데이터 셋 구축
모든 시나리오에 대한 SQLite 샌드박스 파일을 생성합니다.
```powershell
.\scripts\ps1\01-setup-research-data.ps1
```
*   **결과**: `experiments\data\` 폴더에 시나리오별 `.db` 파일이 생성됩니다.

### Step 2: 전수 조사 및 CSV 추출
구축된 데이터 셋과 `experiments/ddl` 폴더의 모든 케이스를 결합하여 분석을 실행합니다.
```powershell
.\scripts\ps1\02-run-research-batch.ps1
```
*   **결과**: `experiments\reports\research_results.csv` 파일에 모든 교차 분석 결과가 누적됩니다.

### Step 3: 데이터 시각화 (Visualization)
제공되는 Python 스크립트를 사용하여 그래프를 생성합니다. (Pandas, Seaborn, Matplotlib 필요)

**A. 리스크 분석 결과 히트맵 생성**
```bash
python tools/visualization/plot_results.py experiments/reports/research_results.csv
```
*   **생성물**: `research_results_analysis.png` (시나리오별 리스크 히트맵, 리스크 등급 분포 등)

**B. 특정 시뮬레이션 부하 프로필 생성**
```bash
# 먼저 특정 시나리오의 지표를 CSV로 추출
go run ./cmd/migraguard simulate --scenario experiments/scenarios/03_spike.yaml --csv spike_metrics.csv --no-db
# 그래프 생성
python tools/visualization/plot_load.py spike_metrics.csv
```
*   **생성물**: `spike_metrics_load.png` (시간 흐름에 따른 TPS 및 커넥션 변화 그래프)

---

## 3. CSV 데이터 구조 상세 (Output Schema)

... (기존 내용 유지) ...

추출된 `research_results.csv` 파일의 각 컬럼이 의미하는 바는 다음과 같습니다.

| 컬럼명 | 설명 | 활용 방안 |
| :--- | :--- | :--- |
| **Timestamp** | 분석이 수행된 시간 | 실험 이력 관리 |
| **Scenario** | 적용된 시뮬레이션 샌드박스 파일명 | 트래픽 패턴(정상/피크 등) 식별 |
| **SQLFile** | 분석 대상 마이그레이션 SQL 파일명 | DDL 종류(안전/위험) 식별 |
| **TableName** | 대상 테이블 명 | 테이블별 부하 비교 |
| **LockLevel** | PostgreSQL 락 레벨 (1~8) | 락 강도에 따른 영향 분석 |
| **RewriteRequired** | 테이블 재작성 발생 여부 | I/O 부하의 핵심 변수 |
| **RiskScore** | 최종 리스크 점수 (%) | 전체 위험도 수치화 |
| **RiskLevel** | 등급 (Safe / Warning / Danger) | 게이트키핑 결과 분류 |
| **T_ddl** | 예상 DDL 수행 시간 (ms) | 작업 시간의 정량적 예측값 |
| **T_block** | 총 서비스 블로킹 시간 (ms) | 장애 지속 시간 예측 |
| **C_peak** | 예측 피크 커넥션 수 | 시스템 붕괴 여부 판단 지표 |
| **T_rec** | 서비스 정상화 회복 시간 (ms) | 장애 복구 탄력성 분석 |
| **BaseTPS** | 가중치가 적용된 기준 TPS ($\lambda$) | 부하 강도의 기준점 |
| **TableSize** | 분석 시점의 테이블 크기 (Bytes) | 데이터 규모에 따른 리스크 변동성 분석 |

---

## 4. 연구 활용 팁 (Research Tips)

### 4.1. 엑셀/Google Sheets 활용
생성된 CSV를 엑셀에서 열어 **피벗 테이블(Pivot Table)**을 생성해 보세요.
*   **시나리오별 평균 리스크 점수**: 어떤 트래픽 상황에서 특정 DDL이 급격히 위험해지는지 차트화할 수 있습니다.
*   **TableSize vs T_ddl 상관관계**: 테이블 크기가 커짐에 따라 리스크가 선형적으로 증가하는지, 지수적으로 증가하는지 분석할 수 있습니다.

### 4.2. Python (Pandas/Matplotlib) 시각화
연구원들은 다음과 같이 데이터를 로드하여 시각화할 수 있습니다.
```python
import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv('experiments/reports/research_results.csv')
# RiskScore가 Danger인 항목들만 필터링하여 분석
danger_cases = df[df['RiskLevel'] == 'Danger']
danger_cases.groupby('Scenario')['RiskScore'].mean().plot(kind='bar')
plt.title('Average Risk Score by Scenario (Danger Cases)')
plt.show()
```

---

## 5. 주의 사항
*   `run_research.bat`는 실행 시 기존 `research_results.csv` 파일을 초기화하고 새로 작성합니다. (데이터 유실 주의)
*   분석할 DDL 파일이 많을 경우 시간이 다소 소요될 수 있으니(Case당 약 0.5초), 진행 상황 메시지를 확인하세요.
