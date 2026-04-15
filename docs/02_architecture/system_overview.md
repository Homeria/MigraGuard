# 🏛️ MigraGuard 시스템 아키텍처 개요 (System Overview)

본 문서는 MigraGuard의 **기능 중심 아키텍처(Feature-First Architecture)**와 데이터 흐름, 그리고 핵심 리스크 모델의 설계 사상을 다룹니다.

---

## 1. 기능 중심 아키텍처 (Feature-First Architecture)

MigraGuard는 코드의 직관성과 도메인 응집도를 높이기 위해 기능을 중심으로 5개의 논리적 레이어로 구성됩니다.

### 1.1. 주요 레이어 구성
- **Analyzer (The Brain)**: SQL 파싱 및 5단계 리스크 엔진을 포함하는 핵심 분석 도메인입니다.
- **Collector (The Ear)**: PostgreSQL의 지표를 수집하고 델타(Delta) 값을 계산하는 데이터 정제 도메인입니다.
- **App (The Orchestrator)**: CLI 명령과 도메인 로직을 연결하는 어플리케이션 서비스 레이어입니다.
- **Infrastructure (The Hands)**: Postgres, SQLite 등 외부 시스템과의 통신 및 설정을 담당하는 어댑터 레이어입니다.
- **Shared (The Foundation)**: 전역에서 사용하는 데이터 모델(Types), 에러 정의, 리포터 등을 포함하는 공용 레이어입니다.

---

## 2. 데이터 인프라 및 공유 전략

| 구성 요소 | 설명 | 위치 |
| :--- | :--- | :--- |
| **PostgreSQL** | 분석 대상이 되는 운영 데이터베이스. | `internal/infra/postgres` |
| **SQLite** | 수집된 메트릭이 저장되는 시계열 데이터베이스. | `internal/infra/sqlite` |
| **Domain Models** | 시스템 전역에서 통용되는 핵심 데이터 구조체. | `internal/shared/types` |

---

## 3. 5단계 정밀 리스크 모델 (5-Step Risk Model)

MigraGuard는 아래의 5단계를 거쳐 최종 리스크 점수를 도출합니다. (상세 로직은 `internal/analyzer/risk.go` 참조)

1.  **Step 1: $T_{ddl}$ (스키마 분석)**: DDL 유형 기반 기본 작업 시간 예측.
2.  **Step 2: $T_{block}$ (락 경합 예측)**: 잠재적 블로킹 시간 산출.
3.  **Step 3: $C_{peak}$ (피크 동시성 분석)**: 최대 유입 커넥션 예측.
4.  **Step 4: $T_{rec}$ (회복 시간 예측)**: 시스템 정상화까지의 복구 비용 평가.
5.  **Step 5: 리스크 등급 확정**: 최종 가중치를 적용하여 Safe / Warning / Danger 판정.

---

## 4. 설계 상세 다이어그램 (Detailed Design Diagrams)

MigraGuard의 정밀한 설계 구조와 데이터 흐름은 다음의 UML 다이어그램 문서에서 확인할 수 있습니다.

- **[전체 지도] [전역 구현 지도 (05.d2)](./uml_diagrams/05.d2)**: **[최신]** 계층형 수직 스택 및 함수 그리드 구조의 통합 설계도.
- **[도메인-분석] [리스크 분석 상세 흐름 (analyze_flow.d2)](./uml_diagrams/analyze_flow.d2)**: SQL 파싱부터 리스크 평가까지의 핵심 분석 파이프라인.
- **[도메인-에이전트] [지표 수집 상세 흐름 (agent_flow.d2)](./uml_diagrams/agent_flow.d2)**: 백그라운드 지표 수집 및 시계열 데이터 영속화 프로세스.
- **[동적 흐름] [리스크 분석 시퀀스 (Analyze Sequence)](./uml_diagrams/02_sequence_analyze.md)**: 분석 파이프라인의 동적 호출 흐름.
- **[판단 로직] [리스크 등급 상태도 (Risk State Diagram)](./uml_diagrams/04_state_risk_engine.md)**: 위험도 판정 및 게이트키핑 상태 변화.
