# 🔮 예측 분석 엔진 (Predictive Analysis Engine): 예측 로직

이 문서는 MigraGuard v3.8에 구현된 향후 24시간 리스크 예측 및 Safe Window 추천 엔진의 기술적 원리를 설명합니다.

---

## 1. 개념: 사후 분석에서 사전 예측으로

기존의 분석이 데이터베이스의 "현재" 상태에 집중했다면, **예측 엔진**은 *"앞으로 24시간 중 이 DDL을 실행하기에 가장 안전한 때는 언제인가?"*라는 질문에 답합니다.

이를 위해 과거 워크로드 패턴을 합성하여 미래 트래픽을 추정하고, 총 24회의 "가상 실험(What-If Simulation)"을 수행합니다.

---

## 2. 기술적 워크플로우

### 2.1. 과거 이력 프로파일링 (집계)
엔진은 로컬 SQLite 메트릭 저장소를 조회하여 베이스라인 프로필을 구축합니다. 최근 7일간의 데이터를 시간대(Hour of day, 00~23시)별로 그룹화하여 통계를 산출합니다.

**SQL 로직**:
```sql
SELECT 
    CAST(substr(timestamp, 12, 2) AS INTEGER) as hour, 
    AVG(tps) as avg_tps, 
    MIN(tps) as min_tps,
    MAX(tps) as max_tps,
    AVG(p99_time) as avg_p99 
FROM table_metrics 
WHERE table_name = ? 
GROUP BY hour;
```

### 2.2. In-Memory "What-If" 시뮬레이션 루프
`RiskEngine`은 단일 계산에 그치지 않고, 예측된 24개의 타임슬롯을 순회합니다.
- **입력**: 매 시간 $H \in \{0..23\}$ 마다, 해당 시간의 예상 TPS와 Latency를 리스크 수학 모델에 주입합니다.
- **최적화**: 테이블 크기와 같은 정적 메트릭은 1회만 조회하며, 동적 변수만 메모리 상에서 교체하여 24회 연산을 매우 빠르게(50ms 미만) 완료합니다.

---

## 3. 시각화: 리스크 히트맵 (Risk Heatmap)

분석 결과는 `predictive_forecast.csv`로 추출되어 Python 기반 시각화 도구(`plot_predictive_heatmap.py`)에 의해 처리됩니다.

### 3.1. 변동성 구름 (Variance Cloud)
예측의 불확실성을 표현하기 위해, 평균 TPS 선 주변에 과거 데이터의 최솟값(Min)과 최댓값(Max) 범위를 연한 그림자(Shaded Area) 형태로 표시합니다.
- **실선**: 예상(평균) 워크로드.
- **음영**: 트래픽의 실제 변동 가능 범위.

### 3.2. 리스크 기반 배경색 (Heatmap Overlay)
그래프 배경은 리스크 점수에 따라 자동으로 색칠되어 직관적인 피드백을 제공합니다:
- 🟢 **Safe Zone**: 예측 리스크 점수 < 50%
- 🟡 **Warning Zone**: 50% ≤ 예측 리스크 점수 < 80%
- 🔴 **Danger Zone**: 예측 리스크 점수 ≥ 80% (또는 커넥션 초과 예상)

---

## 4. 지능형 의사결정 지원 (Decision Support)

엔진은 24시간 중 리스크 점수가 가장 낮은 시간을 **골든 윈도우(Golden Window)**로 선정합니다. 최종 리포트 그래프 상에 별표(★)와 화살표 주석을 추가하여 사용자가 즉각적으로 최적의 작업 시점을 인지할 수 있게 돕습니다.
