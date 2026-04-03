# MigraGuard v3.1 기술 사양 및 아키텍처 (Agent-CLI 분리)

## 1. 이중화 아키텍처 (Dual-Process Architecture)

MigraGuard는 데이터 수집의 지속성과 분석의 즉각성을 보장하기 위해 두 개의 실행 모드로 분리됩니다.

### 1.1. MigraGuard Agent (Background Service)
- **역할:** 운영 데이터베이스의 메트릭을 중단 없이 수집하여 시계열 저장소를 유지함.
- **동작 방식:**
  - 1분(또는 설정된 주기)마다 `pg_stat_statements`를 조회하여 누적 지표를 캡처.
  - SQLite `workload_snapshots` 테이블에 저장.
  - 데이터 보존(Retention): 설정된 기간(예: 7일)이 지난 지표는 자동 삭제하여 저장소 크기 관리.
- **배포:** 운영 DB 환경에 컨테이너로 상시 가동.

### 1.2. MigraGuard Analyze (Foreground CLI)
- **역할:** 개발자의 마이그레이션 SQL을 입력받아 즉각적으로 위험도를 분석함.
- **동작 방식:**
  - `migraguard.db` 파일에 직접 접근하거나(Shared Volume), 추후 Agent API를 통해 데이터를 조회.
  - AST 파싱 결과로 나온 타겟 테이블의 **최근 1시간 평균 TPS, 최근 24시간 최대 TPS** 등을 즉시 산출.
  - `time.Sleep` 없이 즉각적인 리스크 리포트 생성.
- **배포:** CI/CD 파이프라인(GitHub Actions 등)의 단계로 실행.

## 2. 데이터 공유 전략 (Data Sharing)

| 방식 | 설명 | 장점 | 단점 |
| :--- | :--- | :--- | :--- |
| **Docker Volume** | 동일 호스트 내에서 SQLite 파일을 공유 폴더에 저장 | 구현이 매우 단순하고 빠름 | 물리적으로 떨어진 서버 간 공유 어려움 |
| **Sidecar Pattern** | K8s 환경에서 동일 Pod 내에 Agent와 CLI를 배치 | 컨테이너 간 리소스 공유 최적화 | CI/CD 환경에 따라 설정 복잡 |
| **Agent API (v4.0 예정)** | Agent가 HTTP 서버를 띄워 CLI에 JSON으로 지표 전달 | 네트워크 격리 환경에서도 사용 가능 | API 서버 및 보안 구현 필요 |

## 3. 리스크 분석 로직의 고도화 (v3.1)

- **As-Is (v3.0):** 3초간의 실시간 TPS만 사용 ($\lambda_{3s}$).
- **To-Be (v3.1):**
  - **$\lambda_{avg}$ (Average Load):** 최근 1시간 평균 유입량.
  - **$\lambda_{peak}$ (Peak Load):** 최근 24시간 중 피크 타임 유입량.
  - **$\lambda_{curr}$ (Current Load):** 가장 최근 수집된 실시간 유입량.
  - **최종 분석:** $RiskScore$ 계산 시 위 세 가지 지표를 가중치로 결합하여 "지금 배포하는 것이 안전한가?" 뿐만 아니라 "언제 배포하는 것이 가장 안전한가?"를 판단.

## 4. 데이터 보존 정책 (Retention Policy)
- **SQLite VACUUM:** 주기적인 용량 최적화.
- **Purge Query:** `DELETE FROM workload_snapshots WHERE timestamp < datetime('now', '-7 days')`.
