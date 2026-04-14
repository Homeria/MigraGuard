# 🏛️ MigraGuard 시스템 아키텍처 개요 (System Overview)

본 문서는 MigraGuard의 이중 프로세스 구조와 데이터 흐름, 그리고 핵심 리스크 모델의 설계 사상을 다룹니다.

---

## 1. 이중 프로세스 아키텍처 (Dual-Process Architecture)

MigraGuard는 지표 수집의 지속성과 분석의 즉각성을 동시에 보장하기 위해 두 개의 독립된 프로세스로 운영됩니다.

### 1.1. MigraGuard 에이전트 (Background Service)
- **역할**: 대상 데이터베이스로부터 지속적으로 메트릭을 수집하여 시계열 저장소를 유지관리합니다.
- **핵심 동작**:
  - `pg_stat_statements` 및 `pg_stat_user_tables` 뷰를 주기적으로 스캐닝.
  - **델타 계산(Delta Computation)**: 누적 통계치를 시점별 변화량으로 변환하여 실제 부하량을 도출.
  - **영속적 로우 스토리지**: 에이전트 재시작 시 델타 값의 정확성을 위해 마지막 누적치를 별도 테이블에 보관.
  - **데이터 보관 주기 관리**: 설정된 기간(기본 7일)이 지난 데이터를 자동 삭제하고 `VACUUM`을 통해 공간 최적화.

### 1.2. MigraGuard 분석 CLI (Foreground Tool)
- **역할**: 마이그레이션 SQL 파일을 입력받아 즉각적인 리스크 등급을 계산하고 리포팅합니다.
- **핵심 동작**:
  - 공유된 SQLite 파일에 직접 접근하여 최근 1시간 및 24시간 통계 로드.
  - **AST 구문 분석**: SQL 파서를 통해 대상 테이블을 식별하고 변경 유형(Rewrite 여부 등)을 분석.
  - **스키마 검증**: 분석 대상 테이블이 실제 데이터베이스에 존재하는지 사전에 체크.

---

## 2. 데이터 인프라 및 공유 전략

| 구성 요소 | 설명 | 비고 |
| :--- | :--- | :--- |
| **PostgreSQL** | 분석 대상이 되는 운영 데이터베이스. | `pg_stat_statements` 활성화 필수 |
| **Shared SQLite** | 수집된 메트릭이 저장되는 시계열 데이터베이스. | Docker Volume을 통해 에이전트와 CLI가 공유 |
| **영속성 관리** | 컨테이너 재시작 시에도 데이터가 유지되도록 볼륨 전략 수립. | `docker-compose.yml`에 정의 |

---

## 3. 5단계 정밀 리스크 모델 (5-Step Risk Model)

MigraGuard는 아래의 5단계를 거쳐 최종 리스크 점수를 도출합니다.

1.  **Step 1: $T_{ddl}$ (스키마 분석)**: DDL 유형과 테이블 메타데이터를 기반으로 한 기본 작업 시간 예측.
2.  **Step 2: $T_{block}$ (락 경합 예측)**: 작업 유형 및 테이블 크기에 따른 잠재적 블로킹 시간 산출.
3.  **Step 3: $C_{peak}$ (피크 동시성 분석)**: 24시간 히스토리 데이터를 기반으로 한 최대 유입 커넥션 예측.
4.  **Step 4: $T_{rec}$ (회복 시간 예측)**: 롤백 또는 시스템 과부하 시 정상화까지 걸리는 복구 비용 평가.
5.  **Step 5: 리스크 등급 확정**: 최종 가중치($\lambda_{final}$)를 적용하여 Safe / Warning / Danger 판정.

---

## 4. 설계 상세 다이어그램 (Detailed Design Diagrams)

MigraGuard의 정밀한 설계 구조와 데이터 흐름은 다음의 UML 다이어그램 문서에서 확인할 수 있습니다.

- **[구조 설계] [클래스 다이어그램 (Class & Interface Diagram)](./uml_diagrams/01_class_diagram.md)**: 시스템의 정적 구조와 의존성 주입(DI) 아키텍처 정의.
- **[전체 지도] [전역 구현 맵 (Holistic Implementation Map)](./uml_diagrams/05_holistic_implementation_map.md)**: 소스 코드 파일, 핵심 함수, 컴포넌트 간 상호작용을 집대성한 전역 설계도.
- **[동적 흐름] [리스크 분석 시퀀스 (Analyze Sequence)](./uml_diagrams/02_sequence_analyze.md)**: 사용자의 SQL 분석 요청 처리 파이프라인.
- **[동적 흐름] [지표 수집 시퀀스 (Agent Sequence)](./uml_diagrams/03_sequence_agent.md)**: 백그라운드 지표 수집 및 델타 계산 루프.
- **[판단 로직] [리스크 등급 상태도 (Risk State Diagram)](./uml_diagrams/04_state_risk_engine.md)**: 지표에 따른 위험도 판정 및 배포 관문(Gatekeeping) 로직.
