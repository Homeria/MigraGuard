# 🛡️ MigraGuard: 현업자 자문 피드백 및 시스템 한계점 보고서

본 문서는 현업 시니어 백엔드/SRE(Site Reliability Engineer) 엔지니어의 자문 피드백을 바탕으로, 현재 MigraGuard v3.8 모델이 가진 아키텍처 및 실무 적용 측면의 한계점(Flaws)을 분석하고 이를 방어하기 위한 대응 시나리오를 기록합니다.

---

## 1. 📋 현업 개발자 자문 피드백 요약

| 구분 | 자문 내용 | 실무적 평가 |
| :--- | :--- | :--- |
| **주제 선정** | 배포 직전(CI/CD Gate) 트래픽 인지형 서킷 브레이크 | 매우 우수. 정적 린터와 사후 모니터링 사이의 공백을 메우는 훌륭한 접근법 |
| **파서 신뢰성** | `pg_query_go`를 통한 PostgreSQL 공식 AST 파서 채택 | 매우 우수. 정규식 파서의 오탐(False Positive) 리스크를 원천 차단함 |
| **운영 오버헤드** | SQLite 로컬 캐시 및 1분 주기 경량 메트릭 조회 | 우수. 라이브 DB에 가하는 부하를 최소화하려는 설계가 돋보임 |
| **실무 적용성** | 현 상태 그대로 대기업/핀테크 프로덕션 환경에 바로 도입 가능한가? | **불가능**. 망 분리, 휴리스틱 수식의 불확실성, 대체 기술 존재 등의 장벽 존재 |

---

## 2. ⚠️ MigraGuard의 4대 핵심 허점 (Architectural & Algorithmic Flaws)

### ① SQLite 파일 동기화 및 망 분리 장벽 (The Database Sync Problem)
* **허점**: 
  - 에이전트([agent_service.go](file:///home/gyeongho/Github/MigraGuard/cmd/migraguard/agent_service.go))가 프로덕션 서버에서 수집한 메트릭은 해당 장비의 로컬 SQLite 파일(`migraguard.db`)에 쓰입니다.
  - 하지만 실제 배포를 실행하는 CI/CD 러너(GitHub Actions 등)는 임시 가상 컨테이너 환경에서 실행됩니다. 
  - 물리적으로 격리된 서버에 있는 SQLite 파일을 배포 컨테이너가 어떻게 읽을 것인가에 대한 네트워크 동기화 방안이 부재합니다. 보안망 내부의 파일을 매번 SSH나 S3를 통해 외부 컨테이너로 유출하는 방식은 보안 감사(Security Audit)를 통과할 수 없습니다.
* **실무적 해결 방안 (To-Do)**:
  - SQLite 로컬 파일 의존성을 끊고, 중앙 집중형 **HTTP Metrics API Server**를 구축해야 합니다.
  - 혹은 현업 표준 모니터링 시스템인 **Prometheus나 VictoriaMetrics** 등의 TSDB와 연동하여, 분석기가 HTTP 쿼리를 통해 원격으로 트래픽을 조회하도록 수정되어야 합니다.

### ② 휴리스틱 매직 넘버 기반 수식의 신뢰성 결여 (The Magic Number Problem)
* **허점**:
  - 현재 리스크 스코어 계산 수식([risk_evaluator.go](file:///home/gyeongho/Github/MigraGuard/pkg/migraguard/internal/analyzer/risk_evaluator.go))은 `avg_multiplier: 1.2`, `lockImpact: 0.5` 등 임의의 고정 상수(Magic Number)에 의존하고 있습니다.
  - 실제 DB 성능은 테이블 크기가 같더라도 **인덱스의 개수(리라이트 시 모든 인덱스가 재구성됨), TOAST 테이블 저장 데이터 유무, OS 페이지 캐시 적중률, CPU 스케줄러 상태** 등에 의해 완전히 비선형적으로 변합니다.
  - 툴이 수식에 따라 "Safe"라고 예측해서 주간에 배포했다가, 예상치 못한 내부 요인으로 락 경합이 길어져 장애가 터질 경우 책임을 질 수 없습니다.
* **실무적 해결 방안 (To-Do)**:
  - 고정된 상수를 사용하는 대신, 대상 DB의 과거 실제 DDL 수행 기록(로그 및 `pg_stat_statements` 기록)을 역추적하여 시스템 환경별 상수값들을 지속적으로 튜닝하는 **자동 보정(Calibration) 모델**이 추가되어야 합니다.

### ③ 온라인 DDL 도구(Online Schema Change) 대비 경쟁력 부재 (The "Why Bother?" Problem)
* **허점**:
  - 위험한 DDL(예: 테이블 리라이트 유발 타입 변경)이 감지되면 시스템은 단순 차단(Circuit Break) 및 새벽 시간대(Golden Window) 배포를 추천합니다.
  - 그러나 현업에서는 수십 GB가 넘는 테이블의 스키마 변경 시, 무조건 `gh-ost`나 `pt-online-schema-change` 같은 온라인 마이그레이션 도구를 사용해 락 경합 자체를 기술적으로 우회합니다.
  - "위험하니 피하라"는 식의 알림만 주는 도구는 개발 프로세스를 블로킹하여 생산성만 저하시키는 방해물로 인식될 수 있습니다.
* **실무적 해결 방안 (To-Do)**:
  - 단순 차단 및 배포 시간 지연 제안을 넘어, 위험 DDL 감지 시 **온라인 우회 도구 처방전(Prescription)**을 제공해야 합니다.
  - 예: *"해당 ALTER TABLE 쿼리는 리라이트를 유발하므로 pt-osc 도구 활용을 추천하며, 현재 복제 지연이 1.2초이므로 추천 실행 스크립트는 `pt-online-schema-change --max-lag=2s ...` 입니다"* 와 같은 실무 명령어를 제안하도록 기능을 고도화해야 합니다.

### ④ 프로덕션 DB 접근 최소 권한 및 감사(Audit) 부재 (The Privilege & Audit Problem)
* **허점**:
  - 에이전트가 PostgreSQL 메트릭을 수집하기 위해 DB 연결 정보([migraguard.yaml](file:///home/gyeongho/Github/MigraGuard/migraguard.yaml))를 가집니다.
  - 금융권이나 핀테크 서비스에서는 서드파티 툴이 실시간으로 운영 DB에 접속하는 것 자체가 큰 보안 위협입니다. 특히 에이전트가 조작할 수 있는 권한의 한계가 문서화되어 있지 않고 감사 기능이 없으면 도입이 반려됩니다.
* **실무적 해결 방안 (To-Do)**:
  - 에이전트 전용 DB 계정 생성 시 테이블 데이터(DML)에는 접근하지 못하고, 오직 `pg_catalog`와 `pg_stat_statements` 관련 메트릭 뷰만 조회할 수 있는 **최소 권한 가이드(Least Privilege Guide)**를 공식화해야 합니다.
  - 또한 에이전트가 DB에 조회하는 모든 쿼리를 스스로 기록하는 감사 로그(Audit Log) 명세가 필요합니다.

---

## 3. 🎯 캡스톤 발표 및 논문 심사 대응 시나리오

지도교수님 및 심사위원들이 *"현재 만들어둔 것으로는 실제 DevSecOps에 도입하기 어렵지 않겠느냐"* 라고 질문할 때의 모범 답안 가이드라인입니다.

> **[Q] 모의 데이터(Simulation) 기반이어서 실제 운영 환경에 도입하기는 힘들어 보이는데?**
> - **[A]** "그렇지 않습니다. 본 프로젝트의 아키텍처는 **실제 배포용 라이브 모드**와 **논문 실험용 샌드박스 모드**가 완벽히 격리되어 있습니다. 라이브 모드 구동 시 실제 PostgreSQL 서버에 에이전트를 상주시켜 실시간 운영 트래픽 수치(TPS, P99)를 안전하게 수집하는 실무 엔진이 이미 완성되어 작동합니다. 샌드박스 엔진은 오직 학술 논문의 검증(Evaluation)을 위해 일관되고 재현 가능한 통제 환경을 구성할 목적으로 개발된 도구입니다."
>
> **[Q] 어차피 pt-osc 같은 Online DDL 툴을 쓰면 되는데 왜 굳이 이 리스크 게이트키퍼를 도입해야 하는가?**
> - **[A]** "온라인 마이그레이션 도구는 락을 피할 수 있지만, 복제 지연(Replication Lag) 상태에 따라 여전히 서비스 지연을 유발할 수 있으며, 개발자가 실수로 온라인 도구를 거치지 않고 바로 DDL을 배포하는 휴먼 에러(Human Error)를 막지 못합니다. MigraGuard는 개발자의 실수로 들어오는 날것(Raw)의 위험한 DDL을 배포 전 단계에서 강제로 차단하고 올바른 우회 도구 처방전을 제시하는 **최종 안전망(Safety Net)**의 역할을 하기 때문에 상호 보완적으로 필수 도입되어야 합니다."
