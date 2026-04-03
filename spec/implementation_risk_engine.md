# MigraGuard Phase 3: Risk Engine 구현 리포트 (v3.0 고도화)

## 1. 구현 개요
- **목적**: 대기행렬 이론(Queuing Theory)을 기반으로 DDL 실행 시 발생하는 커넥션 급증 및 연쇄 장애 리스크를 정량적으로 산출.
- **핵심 모듈**: `internal/engine/risk.go`

## 2. v3.0 수학적 모델 구현 상세
- **Step 1. $T_{ddl}$ 추정**:
    - 파서의 `RewriteRequired`($F_{rewrite}$) 플래그와 DB의 `TableSize`($S_{table}$)를 결합하여 물리적 소요 시간 계산.
- **Step 2. $T_{block}$ 산출**:
    - $T_{p99}$ (꼬리 지연) + $T_{ddl}$ (실행 시간) + $Lag_{repl}$ (복제 지연)을 합산하여 총 블로킹 시간 도출.
- **Step 3. $C_{peak}$ 산출**:
    - $C_{active} + (\lambda \times T_{block})$ 공식을 통해 락 해제 직후의 최대 커넥션 요구량 예측.
- **Step 4. $T_{rec}$ 및 영구 장애 판단**:
    - 시스템 최대 처리량($\mu_{max}$) 대비 유입량($\lambda$)을 비교하여 회복 시간($T_{rec}$) 산출.
    - $\lambda \geq \mu_{max}$인 경우 `PermanentFailure`로 판별.
- **Step 5. $RiskScore(\%)$ 산출**:
    - $(C_{peak} / C_{max}) \times 100$을 통해 커넥션 고갈 위험도를 점수화.

## 3. 주요 데이터 구조
### `RiskAnalysisReport` 구조체
- `RiskScore`: 최종 위험도 점수.
- `RecoveryTime`: 시스템이 정상 상태로 돌아오는 데 걸리는 시간.
- `RiskLevel`: Danger (90%+), Warning (60%+), Safe 기준 분류.

## 4. 향후 과제
- 중앙값(Median) 기반의 트래픽 밀도 분석을 통한 **Safe Window(안전 배포 시간대)** 추천 로직 추가.
- $Disk_{IO}$, $\mu_{max}$ 등 인프라 상수의 실측 데이터 기반 자동 보정 로직.
