# 🛡️ MigraGuard 프로젝트 진행 상황 (v3.8 시각적 안전 구간 발견 기능 완료)

본 문서는 v3.8 시뮬레이션 환경 구축 이후, 이를 기반으로 향후 24시간의 리스크를 예측하고 시각화하는 지능형 의사결정 지원 시스템 구축 과정을 기록합니다.

## 📅 마지막 업데이트: 2026-05-15 (v3.8-Step 7: Predictive Safe Window Discovery 완성)
- **현재 상태:** **향후 24시간 트래픽 예측 및 시간대별 리스크 히트맵 시각화 파이프라인 구축 완료**
- **핵심 성과:** 
  - **Predictive Engine**: 과거 데이터를 기반으로 24시간 워크로드 프로필을 생성하고, 메모리 상에서 리스크 모델을 24회 반복 시뮬레이션하는 기능 구현.
  - **Risk Heatmap Visualization**: Python Matplotlib을 이용해 예상 TPS 선 그래프 위에 리스크 수준(Safe/Warning/Danger)을 배경색으로 입힌 히트맵 생성.
  - **Traffic Variance Cloud**: 단순 평균선이 아닌 실제 데이터의 Min-Max 범위를 투명한 구름 형태로 시각화하여 예측 신뢰도 향상.
  - **Unified Workflow**: `--forecast` 플래그 하나로 분석부터 그래프 생성(`predictive_risk_heatmap.png`)까지 자동화된 쉘 스크립트 기반 UX 제공.
  - **CI/CD Integrity**: 일반 분석 모드(Circuit Breaker)와 예측 모드(Decision Support)를 철저히 분리하여 아키텍처 정체성 유지.

## 🚀 향후 로드맵 (Phase 20+ 시뮬레이션 고도화 및 적응형 제안)

> **⚠️ 필독: 다음 세션 시작 전 수행 사항**
> 
> **발표용 데모를 위해 `02_sine_daily_cycle` 시나리오와 `011_danger_rewrite` DDL 조합으로 가장 드라마틱한 히트맵 리포트를 미리 생성해 둘 것.**
> 또한, 현재 루트에 생성되는 `predictive_forecast.csv`와 이미지 파일을 `experiments/reports/`로 자동 이동하는 스크립트 보완 필요.

1.  **`feat/realistic-workload-simulation` (NEXT PRIORITY)**:
   - **구현 목표**: 매일 동일한 사인 곡선이 아닌, 요일별 가중치, 피크 시간 무작위 이동, 비대칭 부하 곡선 등을 적용하여 실제 서비스에 근접한 불규칙한 데이터 생성.
   - **기대 효과**: 히트맵의 변동성 구름(Min-Max)이 훨씬 역동적으로 표현되어 연구적 가치 증대.

2. **`feat/adaptive-recommendation`**:
   - **구현 목표**: 수집된 200개 사례 데이터를 분석하여 시스템 환경별 최적 임계값(`mu_max`, `disk_io`)을 머신러닝/통계 기반으로 자동 제안.


---
**세션 종료:** 이제 MigraGuard는 "장애를 막는 방패"를 넘어 "최적의 경로를 안내하는 나침반"의 기능을 갖췄습니다. 캡스톤 디자인의 핵심인 시각적 결과물(Heatmap)은 이제 명령어 한 줄로 즉시 생성 가능합니다.

**사용자의 메모:** 예측 엔진 안정화 완료. 변동성 시각화(Cloud) 도입으로 전문성 확보. 다음 단계는 시뮬레이터의 '무작위성'을 강화하여 데이터 리얼리티를 높이는 것임.