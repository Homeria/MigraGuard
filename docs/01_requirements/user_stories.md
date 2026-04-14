# 🛡️ MigraGuard 요구사항 명세서 (Requirements Specification)

**버전:** 3.5.0 (Adaptive Config 설계 단계)  
**최근 수정일:** 2026-04-14  
**프로젝트 성격:** PostgreSQL 마이그레이션 리스크 분석 및 자동 게이트키핑 도구

---

## 1. User Stories

### Epic 1: 마이그레이션 리스크 분석 (Analyze)
**Goal:** 개발자가 SQL 배포 전, 운영 환경의 부하를 반영한 정량적 위험도를 확인하고 장애를 예방한다.

**US-01 : 사용자는 특정 SQL 파일의 마이그레이션 리스크를 분석할 수 있다.**
- **Description:**
  - **As a :** 서비스 개발자
  - **I want to :** 배포 예정인 SQL 파일을 CLI 도구에 입력하여 분석을 실행하여
  - **So that :** 해당 쿼리가 운영 환경에서 락(Lock) 경합으로 인한 장애를 유발하지 않도록 사전 예방하고 싶다.
- **Acceptance Criteria:**
  - `analyze <file_path>` 명령을 통해 분석이 시작되어야 한다.
  - SQL 파서를 통해 테이블 재작성(Table Rewrite) 발생 여부를 정확히 감지해야 한다.
  - 분석 결과는 테이블 형태의 리포트로 출력되어야 하며, 최종 리스크 등급(Safe/Warning/Danger)이 명시되어야 한다.

**US-02 : 사용자는 분석 결과 리포트를 마크다운 형식으로 추출할 수 있다.**
- **Description:**
  - **As a :** 서비스 개발자
  - **I want to :** 분석 결과를 마크다운 포맷으로 저장하여
  - **So that :** GitHub Pull Request의 코멘트로 활용하거나 동료들과 공유할 수 있다.
- **Acceptance Criteria:**
  - `--format markdown` 플래그 지원을 통해 표준 마크다운 문법의 텍스트를 출력해야 한다.
  - 리포트에는 예상 블로킹 시간($T_{block}$)과 기준 TPS 정보가 포함되어야 한다.

---

### Epic 2: 운영 지표 수집 및 관리 (Agent)
**Goal:** 운영 데이터베이스의 트래픽 패턴을 지속적으로 학습하여 분석의 근거 데이터를 확보한다.

**US-03 : 사용자는 백그라운드에서 실행되는 지표 수집 에이전트를 가동할 수 있다.**
- **Description:**
  - **As a :** DBA / 시스템 운영자
  - **I want to :** MigraGuard 에이전트를 상시 실행하여
  - **So that :** 운영 DB의 실시간 및 과거 피크 트래픽 데이터를 지속적으로 누적할 수 있다.
- **Acceptance Criteria:**
  - `agent` 명령을 통해 백그라운드 수집 프로세스가 시작되어야 한다.
  - 에이전트는 설정된 주기(기본 1분)마다 DB 메트릭을 수집해야 한다.
  - 에이전트 종료 후 재시작 시에도 데이터 연속성이 보장되어야 한다.

**US-04 : 사용자는 수집된 지표를 바탕으로 안전한 배포 시간대를 추천받을 수 있다.**
- **Description:**
  - **As a :** 서비스 개발자
  - **I want to :** 최근 24시간의 트래픽 추이를 분석한 추천 데이터를 확인하여
  - **So that :** 트래픽이 집중되는 시간을 피해 마이그레이션을 안전하게 수행할 수 있는 골든 타임을 결정할 수 있다.
- **Acceptance Criteria:**
  - 분석 리포트 하단에 "Recommended Window" 섹션이 포함되어야 한다.
  - 최근 24시간 중 TPS가 가장 낮은 시간대를 계산하여 제시해야 한다.

---

### Epic 3: 적응형 설정 및 지능형 추천 (v3.5 Adaptive)
**Goal:** 시스템이 스스로 인프라 환경을 분석하여 리스크 판단 기준을 최적화한다.

**US-05 : 사용자는 현재 시스템 환경에 최적화된 리스크 상수를 추천받을 수 있다.**
- **Description:**
  - **As a :** DBA
  - **I want to :** 실제 수집된 데이터에 기반한 권장 리스크 설정값(MuMax, DiskIO 등)을 확인하여
  - **So that :** 우리 인프라 성능에 딱 맞는 정밀한 리스크 분석 환경을 구축할 수 있다.
- **Acceptance Criteria:**
  - 과거 피크 트래픽과 디스크 응답 속도를 분석하여 적정 상수값을 도출해야 한다.
  - 현재 설정값과 추천값을 비교하여 차이가 클 경우 사용자에게 알림을 제공해야 한다.

---

## 2. System Stories (Technical Specification)

### Epic 1: SQL 구문 분석 및 리스크 엔진
**Goal:** SQL의 정적 특성과 운영 지표를 결합한 수학적 모델링을 수행한다.

**SYS-01 : PostgreSQL AST 파싱 및 DDL 특성 추출**
- **Description:**
  - **As a :** 시스템
  - **I want to :** `pg_query_go`를 사용하여 SQL을 추상 구문 트리로 변환하고 분석하여
  - **So that :** 대상 테이블 이름, 컬럼 변경 사항, Table Rewrite 유발 여부를 판별할 수 있다.
- **Acceptance Criteria:**
  - `ALTER TABLE` 명령 시 `SET DEFAULT`, `TYPE CHANGE` 등 Rewrite 여부를 분석해야 한다.
  - 분석 대상 테이블이 실제 DB에 존재하는지 스키마 검증을 수행해야 한다.

**SYS-02 : 5단계 정밀 리스크 산출 알고리즘 적용**
- **Description:**
  - **As a :** 리스크 엔진
  - **I want to :** 대기 행렬 이론을 적용한 5단계 수식을 통해 리스크 점수를 계산하여
  - **So that :** 운영 환경에서 발생할 수 있는 잠재적 장애 가능성을 보수적으로 예측할 수 있다.
- **Acceptance Criteria:**
  - $T_{ddl}$(예상 작업시간), $T_{block}$(블로킹 시간), $C_{peak}$(피크 커넥션) 등을 순차적으로 계산해야 한다.
  - $\lambda_{final} = \max(\lambda_{curr}, \lambda_{avg\_1h} \times 1.2, \lambda_{peak\_24h} \times 0.8)$ 공식을 적용해야 한다.

---

### Epic 2: 메트릭 수집 및 저장소 관리
**Goal:** 데이터의 무결성을 유지하며 효율적인 시계열 데이터 저장소를 운영한다.

**SYS-03 : 시계열 데이터 델타(Delta) 계산 및 적재**
- **Description:**
  - **As a :** 에이전트 서비스
  - **I want to :** Postgres의 누적 통계치에서 이전 값을 차감하여 순수 변화량(Delta)을 계산하여
  - **So that :** 특정 수집 주기의 실제 TPS와 부하량을 정확히 기록할 수 있다.
- **Acceptance Criteria:**
  - SQLite 내에 `original_pg_stat_statements` 테이블을 두어 마지막 누적치를 관리해야 한다.
  - 계산된 델타 값은 `workload_snapshots` 테이블에 시계열로 저장해야 한다.

**SYS-04 : 자동 가비지 컬렉션 (Retention)**
- **Description:**
  - **As a :** 백그라운드 스케줄러
  - **I want to :** 설정된 보관 주기(기본 7일)가 지난 데이터를 물리적으로 삭제하여
  - **So that :** 로컬 저장소의 용량 포화로 인한 시스템 마비를 방지할 수 있다.
- **Acceptance Criteria:**
  - 매 수집 주기마다 `DELETE` 쿼리를 실행하여 만료된 데이터를 삭제해야 한다.
  - 삭제 후 `VACUUM` 명령을 실행하여 데이터베이스 파일을 최적화해야 한다.

---

### Epic 3: 배포 관문(Gatekeeping) 및 통합
**Goal:** 위험한 배포를 시스템 수준에서 차단하고 외부 도구와 연동한다.

**SYS-05 : 비정상 종료를 통한 CI/CD 게이트키핑**
- **Description:**
  - **As a :** CLI 어플리케이션
  - **I want to :** 분석 결과 리스크 등급이 `Danger`일 경우 프로세스 종료 코드를 `1`로 반환하여
  - **So that :** GitHub Actions 등의 CI 도구에서 후속 배포 단계가 진행되지 않도록 강제 차단할 수 있다.
- **Acceptance Criteria:**
  - 분석 완료 시 리스크 등급에 따라 `os.Exit(0)` 또는 `os.Exit(1)`을 호출해야 한다.

---

## 3. Acceptance Criteria (공통 품질 기준)

- **정확성:** 10만 건 이상의 대형 테이블에서 Table Rewrite가 발생하는 모든 DDL을 'Danger'로 탐지해야 함.
- **성능:** SQL 분석 단계의 소요 시간은 로컬 DB 조회 포함 3초 이내여야 함.
- **호환성:** PostgreSQL 14 이상의 버전과 공식 호환되며 `pg_stat_statements` 확장과 연동되어야 함.
- **안정성:** 에이전트 가동 시 운영 데이터베이스의 CPU 점유율 상승폭이 1.0%p 이내여야 함.
