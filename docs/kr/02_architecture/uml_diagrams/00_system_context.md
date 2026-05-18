# Level 0: 시스템 컨텍스트 다이어그램 (System Context)

이 다이어그램은 데이터베이스 마이그레이션 라이프사이클 동안 **MigraGuard**가 외부 액터 및 시스템 컴포넌트와 어떻게 상호작용하는지에 대한 고수준 조감도를 제공합니다.

```mermaid
graph TD
    subgraph Users [사용자 레이어]
        DEV[개발자]
        CICD[CI/CD 파이프라인]
    end

    subgraph MigraGuard_Ecosystem [MigraGuard 에코시스템]
        CLI[MigraGuard CLI]
        SDK[Core SDK - pkg/migraguard]
        PY[Python 시각화 도구 - Matplotlib]
    end

    subgraph Data_Layers [데이터 레이어]
        PG[(대상 PostgreSQL)]
        SQLITE[(샌드박스/메트릭 SQLite)]
    end

    DEV -->|명령어 실행| CLI
    CICD -->|게이트키퍼/분석 실행| CLI
    
    CLI --> SDK
    SDK -->|실시간 메트릭 수집| PG
    SDK -->|이력 저장 및 조회| SQLITE
    SDK -->|샌드박스 DB 생성| SQLITE
    
    CLI -->|CSV 데이터 내보내기| PY
    PY -->|히트맵 PNG 생성| CICD
    PY -->|히트맵 PNG 생성| DEV

    style CLI fill:#f9f,stroke:#333,stroke-width:4px
    style SDK fill:#bbf,stroke:#333,stroke-width:2px
    style PG fill:#3498db,stroke:#333,stroke-width:2px
    style SQLITE fill:#95a5a6,stroke:#333,stroke-width:2px
```

### 주요 상호작용
1.  **개발자/파이프라인**: CLI를 통해 프로세스를 시작합니다 (예: `analyze`, `simulate`).
2.  **CLI & SDK**: CLI는 수학적 리스크 엔진을 포함하고 있는 Core SDK의 래퍼(Wrapper) 역할을 합니다.
3.  **데이터 지속성**: 실시간 메트릭은 PostgreSQL에서 수집되며, 과거 트렌드 및 시뮬레이션 데이터는 SQLite에 저장됩니다.
4.  **시각화**: Python 스크립트는 분석 데이터를 소비하여 고해상도 리포트를 생성하는 다운스트림 역할을 수행합니다.
