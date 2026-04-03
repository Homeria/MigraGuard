# MigraGuard v3.1 상세 로직 명세서 (Agent-CLI 분리 모델)

본 문서는 에이전트가 상시 수집한 데이터를 바탕으로 CLI가 분석을 수행하는 v3.1 로직의 최종 구현 상태를 정의합니다.

---

## 🛰️ Phase A: Agent - 상시 지표 수집 (Continuous Collection)
**파일:** `cmd/migraguard/agent.go`, `internal/db/collector.go`, `internal/db/activity.go`

- **[L-A01] 에이전트 독립 실행**: `migraguard agent` 커맨드를 통해 백그라운드 프로세스로 상시 가동.
- **[L-A02] 주기적 스냅샷 캡처**: 정해진 주기(`--interval`)마다 `pg_stat_statements` 및 테이블별 동적 지표를 수집하여 SQLite에 저장.
- **[L-A03] 데이터 보존 정책 (Retention)**: 수집 주기마다 `PurgeOldSnapshots()`를 호출하여 설정된 기간(`--retention`, 기본 7일)이 지난 데이터를 자동 삭제하고 `VACUUM`으로 최적화.
- **[L-A04] Graceful Shutdown**: OS 시그널 감지 시 현재 수집 루프를 안전하게 마치고 종료.

---

## 🔍 Phase B: CLI - 즉각 리스크 분석 (Instant Analysis)
**파일:** `cmd/migraguard/analyze.go`, `internal/engine/risk.go`

- **[L-B01] SQL 로드 및 정적 분석**: 마이그레이션 SQL을 파싱하여 타겟 테이블 및 DDL 특성($F_{rewrite}$) 파악.
- **[L-B02] 시계열 지표 즉시 활용**: 
  - 기존의 3초 대기(`time.Sleep`) 로직을 완전히 제거.
  - 리스크 엔진 호출 시 SQLite에 축적된 최근 스냅샷들로부터 TPS($\lambda$) 델타값을 즉각 계산.
- **[L-B03] 리스크 산출**: 
  - SQLite 델타 TPS를 최우선으로 사용하며, 데이터가 없을 경우에만 Postgres의 누적 평균치로 폴백(Fallback).
  - 큐잉 모델 수식을 통해 $T_{block}$, $C_{peak}$, $RiskScore$ 도출.
- **[L-B04] 게이트키핑**: 분석 결과가 `Danger` 레벨일 경우 `Exit Code 1`을 반환하여 CI/CD 파이프라인 차단.

---

## 🧠 Phase C: 리스크 엔진 수식 (v3.1 Baseline Model)

**핵심 수식:**
- **TPS 산출**: $\lambda = \frac{\Delta Calls}{\Delta Time}$ (SQLite 최근 2개 스냅샷 기준)
- **큐 스파이크**: $C_{peak} = C_{active} + (\lambda \times T_{block})$
- **위험도 점수**: $RiskScore(\%) = (C_{peak} / C_{max}) \times 100$

---

## 🚦 Phase D: 배포 및 운영 가이드 (DevSecOps)
- **에이전트 배포**: 운영 DB 네트워크 내에서 볼륨 공유가 가능한 형태로 상시 가동.
- **분석 실행**: CI/CD Runner가 에이전트의 SQLite 파일에 접근하여 `analyze` 명령 실행.
