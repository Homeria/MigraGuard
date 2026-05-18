# Level 1: 명령어 단위 워크플로우 (Mid-Level)

이 문서는 주요 MigraGuard 명령어의 운영 흐름을 설명합니다.

## 1. `analyze` 워크플로우 (리스크 평가 및 예측)

`analyze` 명령어는 시스템의 핵심 "서킷 브레이커" 역할을 합니다.

```mermaid
flowchart TD
    Start([시작]) --> Parse[SQL 파일 파싱 - AST 분석]
    Parse --> Mode{분석 모드?}
    
    Mode -- Live --> PG[운영 PostgreSQL에서 메트릭 수집]
    Mode -- Sandbox --> SL[SQLite 샌드박스에서 메트릭 수집]
    
    PG & SL --> Engine[5단계 리스크 엔진 실행]
    
    Engine --> Forecast{--forecast 활성화?}
    
    Forecast -- 아니오 --> Report[콘솔/CSV 리포트 생성]
    
    Forecast -- 예 --> History[SQLite에서 24시간 트래픽 베이스라인 추출]
    History --> Loop[24시간 반복 - 리스크 시뮬레이션]
    Loop --> TieBreak[골든 윈도우 식별 - 최저 리스크 및 TPS 기준]
    TieBreak --> Visual[Python: 히트맵 PNG 생성]
    Visual --> Report
    
    Report --> Decision{리스크 > 임계치?}
    Decision -- Danger --> Block([파이프라인 차단 - Exit 1])
    Decision -- Safe --> Allow([파이프라인 승인 - Exit 0])
```

## 2. `simulate` 워크플로우 (연구용 샌드박스 생성)

연구자가 선언적 YAML을 통해 제어된 실험 환경을 구축할 수 있게 합니다.

```mermaid
flowchart TD
    Start([시작]) --> YAML[시나리오 YAML 읽기]
    YAML --> Prof[워크로드 프로파일러 초기화]
    Prof --> DB[새로운 {Experiment}.db 생성]
    DB --> Seed[Sine/Noise 패턴을 가진 7일치 이력 시딩]
    Seed --> Finish([샌드박스 준비 완료])
```

## 3. `agent` 워크플로우 (백그라운드 데이터 수집)

시스템이 예측 분석을 위한 충분한 과거 데이터를 확보하도록 보장합니다.

```mermaid
flowchart TD
    Start([시작]) --> Loop[인터벌 주기 - 예: 1분마다]
    Loop --> Harvest[pg_stat_statements 및 테이블 크기 수집]
    Harvest --> Save[SQLite 메트릭 테이블에 UPSERT]
    Save --> Purge[보관 기간이 지난 데이터 정리]
    Purge --> Loop
```
