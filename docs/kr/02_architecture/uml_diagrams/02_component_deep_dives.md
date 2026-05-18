# Level 2: 핵심 컴포넌트 상세 (Deep-Dives)

이 문서는 리스크 엔진의 내부 로직과 골든 윈도우 탐색 알고리즘을 상세히 다룹니다.

## 1. 5단계 정량적 리스크 모델 (5-Step Model)

각 DDL은 최종 리스크 점수를 산출하기 위해 5가지 체계적인 단계를 거쳐 평가됩니다.

```mermaid
sequenceDiagram
    participant P as PostgreSQL/Sandbox
    participant E as 리스크 엔진
    participant R as 분석 리포트

    E->>P: 메트릭 수집 (TPS, P99, TableSize, ActiveConns)
    Note over E: Step 1: T_ddl (실행 시간 예측)
    E->>E: 디스크 I/O 대비 테이블 크기로 실행 시간 예측
    
    Note over E: Step 2: T_block (락 차단 시간)
    E->>E: T_block = (P99 + T_ddl + 복제지연) * 락가중치
    
    Note over E: Step 3: C_peak (최대 커넥션 폭증)
    E->>E: C_peak = 현재커넥션 + (TPS/1000 * T_block)
    
    Note over E: Step 4: T_rec (장애 회복 시간)
    E->>E: T_rec = (C_peak - 최대수용량) / (한계처리량 - TPS)
    
    Note over E: Step 5: 리스크 점수 및 등급
    E->>E: 점수 = (C_peak / 최대수용량) * 100
    E->>R: 리포트 생성 (Safe/Warning/Danger)
```

## 2. 골든 윈도우 탐색 로직 (Predictive Forecast)

`--forecast` 옵션 사용 시, 엔진은 가장 안전한 배포 시간을 찾기 위해 24번의 가상 시뮬레이션을 수행합니다.

```mermaid
flowchart TD
    Start[24시간 예측 프로필 로드] --> Loop{{각 시간대별 반복 00..23}}
    Loop --> Virtual[가상 메트릭 스냅샷 생성]
    Virtual --> Run[해당 시간대에 대해 5단계 모델 실행]
    Run --> Score[리스크 점수 및 예상 TPS 산출]
    
    Score --> Better{현재 시간대가 더 안전한가?}
    
    Better -- "점수 < 최저점수" --> Update[최적시간 갱신]
    Better -- "점수 == 최저점수 AND TPS < 최저TPS" --> Update
    Better -- 아니오 --> Next[다음 시간대로 이동]
    
    Update --> Next
    Next --> Loop
    Loop -- 종료 --> Result([최종 골든 윈도우 결정])
```

### TPS Tie-Breaker 로직 (v3.8 핵심 개선)
만약 여러 시간대가 동일하게 "안전(Safe)"하다고 판단될 경우(예: 모두 리스크 하한선인 30.0점에 도달), 엔진은 **예상 TPS가 가장 낮은 시간대**를 선택합니다. 이는 물리적인 트래픽이 가장 적은 시점에 작업을 수행하게 함으로써 안전 마진을 극대화합니다.
