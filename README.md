# 🛡️ MigraGuard (DB Migration Gatekeeper)

> **"트래픽을 모르는 DDL은 시한폭탄과 같다."**
> 운영 DB의 실제 쿼리 워크로드(`pg_stat_statements`)와 마이그레이션 스크립트(AST)를 교차 검증하여, 락(Lock) 경합으로 인한 서비스 장애를 배포 이전에 원천 차단하는 DevSecOps CLI 도구.

## 📌 프로젝트 개요 (Project Scope)

본 프로젝트는 데이터베이스 스키마 변경 시 발생할 수 있는 락(Lock) 대기 및 커넥션 풀 고갈 장애를 방지하기 위해 기획되었습니다. 최종 상용화 버전(Full Version)의 비전을 달성하기 위한 첫 단계로서, 캡스톤 디자인의 목적에 맞춘 **간이 버전(Lite/MVP Version)**을 우선적으로 구현합니다.

### 🎯 Full Version vs Lite Version (캡스톤 구현 범위)

| 구분 | 🚀 Full Version (최종 목표) | 🛠️ Lite Version (캡스톤 MVP 구현 범위) |
| :--- | :--- | :--- |
| **데이터 수집** | 백그라운드 데몬 기반 1시간 주기 시계열 스냅샷 자동 누적 | **CLI 실행 시점** 기준 `pg_stat_statements` 단발성/단기 스냅샷 조회 |
| **AST 파싱** | 모든 복잡한 DDL, DML 서브쿼리 완벽 분석 및 의존성 추적 | `ALTER TABLE`, `DROP` 등 **대표적인 배타적 락(Lock) 유발 DDL 위주** 파싱 |
| **위험도 평가** | 중앙값(Median) 연산 기반 **'최적의 안전 시간대(Safe Window)'** 정밀 추천 | 대상 테이블의 **현재 트래픽(TPS) 및 락 레벨 기반 '위험/경고/안전'** 3단계 판정 |
| **이중 검증** | `EXPLAIN` 기반 실행 계획 변화 추적 및 `pg_stat_activity` 실시간 모니터링 연동 | **`pg_stat_activity`를 통한 배포 직전(Pre-flight) 활성 트랜잭션 수 확인** |
| **CI/CD 연동** | GitHub App 봇 연동 (PR에 인터랙티브 차트 및 마크다운 자동 코멘트) | **터미널 표준 출력(Table UI) 및 실패 시 `Exit 1` 반환**을 통한 CI 자동 차단 |
| **상태 저장** | 내장 SQLite 기반 WAL 모드 시계열 데이터 영구 저장소 | 별도 영구 저장 없이 인메모리(In-memory) 연산 후 즉시 결과 반환 |

---

## ✨ 간이 버전(Lite Version) 주요 기능 명세

**1. 마이그레이션 DDL 정적 분석 (AST Parsing)**
* `pganalyze/pg_query_go`를 활용하여 대상 `.sql` 파일을 추상 구문 트리(AST)로 변환.
* 변경이 일어나는 타겟 `Table` 및 `Column` 식별.
* DDL 종류에 따른 요구 Lock Level (예: `AccessExclusiveLock`) 추출.

**2. 런타임 트래픽 조회 (Workload Analysis)**
* 운영 중인 PostgreSQL의 `pg_stat_statements` 뷰를 조회하여, 1단계에서 식별된 타겟 테이블을 참조하는 쿼리들의 실행 빈도(Calls)와 평균 실행 시간(Mean Time) 계산.

**3. CI/CD 게이트키퍼 (Gatekeeping & Pre-flight Check)**
* 산출된 위험도(Risk Score)가 임계치를 초과할 경우 프로세스 종료 코드 `Exit 1`을 반환하여 GitHub Actions 등의 배포 파이프라인 강제 중단.
* 분석 통과 후 실제 배포 스크립트가 돌기 직전, `pg_stat_activity`를 1회 조회하여 해당 테이블을 잡고 있는 장기 실행 쿼리(Long-running Query)가 있는지 최종 확인(Pre-flight Check).

## 🛠️ 기술 스택 (Lite Version 기준)
* **Language:** Go (Golang)
* **Parser:** `pganalyze/pg_query_go` (PostgreSQL C 파서 포팅)
* **DB Driver:** `jackc/pgx` (PostgreSQL 통계 뷰 직접 통신)
* **CLI Framework:** `spf13/cobra` (직관적인 터미널 명령어 지원)

## 🚀 Usage (CI/CD 파이프라인 적용 예시)

GitHub Actions 파일 (`.github/workflows/deploy.yml`) 내에 단일 실행 파일로 손쉽게 통합할 수 있습니다.

```yaml
steps:
  - name: Checkout code
    uses: actions/checkout@v3

  - name: Run MigraGuard (Lite)
    env:
      DB_URL: ${{ secrets.PROD_DB_URL }}
    run: |
      ./migraguard analyze ./migrations/V2__add_email_column.sql
      # 결과가 위험(Danger) 수준일 경우 Exit 1을 반환하여 아래 배포 스텝을 차단함

  - name: Apply Migration
    run: |
      flyway migrate -url=...