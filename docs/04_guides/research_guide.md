# 📊 MigraGuard 연구 데이터 추출 및 분석 가이드 (Research Guide)

본 문서는 MigraGuard의 **일괄 배치 분석(Batch Research)** 기능을 통해 대규모 실험 데이터를 CSV로 추출하고, 이를 논문이나 보고서 작성을 위한 통계 자료로 활용하는 방법을 안내합니다.

---

## 1. 개요 (Overview)

MigraGuard는 단순한 리스크 탐지 도구를 넘어, DDL과 트래픽 간의 상관관계를 정량적으로 연구할 수 있는 실험 플랫폼을 제공합니다. 

*   **전수 조사(Exhaustive Analysis)**: 모든 시뮬레이션 시나리오(11종)에 대해 모든 테스트 DDL(20종)을 교차 분석하여 총 220개 이상의 정밀 데이터를 한 번에 확보합니다.
*   **정량적 지표 추출**: 5단계 리스크 모델의 모든 중간 산출물($T_{ddl}, T_{block}, C_{peak}, T_{rec}$)을 소수점 단위까지 CSV로 추출합니다.

---

## 2. 데이터 확보 프로세스 (Workflow)

윈도우 CMD 환경을 기준으로 설명합니다. (PowerShell 및 Linux용 스크립트도 동일하게 작동합니다.)

### Step 1: 시뮬레이션 데이터 셋 구축
먼저 모든 시나리오에 대한 SQLite 샌드박스 파일을 생성합니다.
```cmd
scripts\cmd\seed_all.bat
```
*   **결과**: `experiments\data\` 폴더에 `steady_normal.db`, `spike_flash_sale.db` 등의 파일이 가득 차게 됩니다.

### Step 2: 전수 조사 및 CSV 추출
구축된 데이터 셋과 테스트 DDL 케이스들을 모두 결합하여 대규모 분석을 실행합니다.
```cmd
scripts\cmd\run_research.bat
```
*   **동작**: 모든 DB 파일과 SQL 파일을 순회하며 분석을 수행하고, 결과를 `experiments\reports\research_results.csv`에 누적합니다.

---

## 3. CSV 데이터 구조 상세 (Output Schema)

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
