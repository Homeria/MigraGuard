# 🚦 리스크 엔진 상태 다이어그램 (Risk Engine State Diagram)

본 다이어그램은 리스크 엔진이 입력받은 지표를 분석하여 최종 위험 등급(Safe, Warning, Danger)으로 전이되는 로직과 판단 기준을 나타냅니다.

## 1. 설계 의도
- **보수적 판정(Conservative Gating)**: 시스템 한계에 근접할수록 급격하게 위험도가 올라가는 상태 변화를 시각화했습니다.
- **장애 인지**: 시스템 마비(Permanent Failure) 상태를 별도로 정의하여 단순한 위험을 넘어선 치명적인 장애를 구분합니다.

## 2. 다이어그램 (Mermaid)

```mermaid
stateDiagram-v2
    [*] --> InitialAnalysis: DDL 분석 요청
    
    InitialAnalysis --> MetricsLoading: 지표 로드 (Postgres/SQLite)
    
    state MetricsLoading {
        [*] --> LambdaCalculation: TPS 산출 (보수적 추정)
        LambdaCalculation --> Step1_5_Evaluation: 5단계 수식 대입
    }
    
    Step1_5_Evaluation --> RiskGrading
    
    state RiskGrading {
        [*] --> DecisionPoint
        DecisionPoint --> Safe: RiskScore < 60%
        DecisionPoint --> Warning: 60% <= RiskScore < 90%
        DecisionPoint --> Danger: RiskScore >= 90%
        DecisionPoint --> SystemOverload: TPS >= Mu_max
    }
    
    Safe --> [*]: 분석 종료 (Pass)
    Warning --> [*]: 분석 종료 (Notice)
    Danger --> DeploymentBlock: 배포 차단 (Exit 1)
    SystemOverload --> DeploymentBlock: 시스템 마비 감지 (Exit 1)
    
    DeploymentBlock --> [*]

    note right of RiskGrading
        RiskScore = (C_peak / C_max) * 100
        C_peak: DDL 수행 중 예측 최대 커넥션
    end
```

## 3. 등급별 상태 정의
- **Safe (안전)**: 예상 부하가 시스템 자원의 60% 미만을 차지하여 마이그레이션을 안전하게 수행할 수 있는 상태.
- **Warning (주의)**: 부하가 60%~90% 사이에 도출되어 잠재적인 지연이 발생할 수 있는 상태. 개발자의 확인이 필요함.
- **Danger (위험)**: 부하가 90%를 초과하여 커넥션 고갈이나 서비스 장애가 확실시되는 상태. 배포 파이프라인 자동 중단.
- **System Overload (시스템 마비)**: 유입 트래픽 자체가 시스템의 최대 처리량(MuMax)을 초과하여 복구가 불가능한 상태.
