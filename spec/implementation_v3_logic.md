# MigraGuard v3.1 상세 로직 명세서 (Agent-CLI 분리 모델)

본 문서는 에이전트가 상시 수집한 데이터를 바탕으로 CLI가 분석을 수행하는 v3.1 로직을 정의합니다.

---

## 🛰️ Phase A: Agent - 상시 지표 수집 (Continuous Collection)
**파일:** `cmd/migraguard/agent.go`, `internal/db/collector.go`

- **[L-A01] 에이전트 초기화**: DB 연결 및 SQLite 저장소 준비.
- **[L-A02] 주기적 스냅샷 캡처**: 정해진 주기(1s~1m)마다 `pg_stat_statements` 데이터를 SQLite에 `INSERT`.
- **[L-A03] 데이터 정제 (Cleanup)**: 매시간 단위로 설정된 보존 기간(Retention)을 초과한 데이터 삭제.
- **[L-A04] 헬스체크**: 수집 상태를 로그로 기록하여 정상 작동 여부 모니터링.

---

## 🔍 Phase B: CLI - 리스크 분석 (On-Demand Analysis)
**파일:** `cmd/migraguard/analyze.go`, `internal/engine/risk.go`

- **[L-B01] SQL 로드 및 파싱**: 사용자가 제공한 DDL 파싱 및 대상 테이블($T_{target}$) 추출.
- **[L-B02] 시계열 지표 쿼리**: **v3.1 핵심.**
  - SQLite에서 $T_{target}$에 대한 최근 1시간 TPS 데이터 조회.
  - `avg()`, `max()` 쿼리를 통해 베이스라인($\lambda_{baseline}$) 산출.
- **[L-B03] 리스크 점수 계산**:
  - `AnalyzeRisk()` 함수에서 `time.Sleep` 없이 즉각적으로 리스크 점수($C_{peak} / C_{max}$) 도출.
  - 만약 SQLite에 데이터가 부족할 경우(신규 에이전트), 경고와 함께 최소 수집 대기 안내.
- **[L-B04] 리포팅**: 현재 트래픽 상황과 대조하여 Safe Window(안전 배포 시간대) 추천 포함.

---

## 🧠 Phase C: 리스크 엔진 수식 (v3.1 Baseline Model)

**기존 수식:** $C_{peak} = C_{active} + (\lambda_{delta} \times T_{block})$

**v3.1 개선 수식:**
- **Weighted TPS ($\lambda_{final}$)**:
  $$\lambda_{final} = (w_1 \times \lambda_{current}) + (w_2 \times \lambda_{avg\_1h}) + (w_3 \times \lambda_{peak\_24h})$$
  *(가중치는 기본적으로 실시간 데이터에 높게 부여하지만, 전체적인 트래픽 추세를 반영함)*

- **Safe Window Recommendation**:
  - 하루 24시간 중 $\lambda_{avg}$가 가장 낮은 1시간 구간을 식별하여 사용자에게 추천.

---

## 🚦 Phase D: 게이트키핑 및 리포팅
- **[L-D01] 결과 시각화**: 현재 리스크 수준과 함께 "최근 24시간 중 트래픽이 높은 시점입니다" 등의 컨텍스트 제공.
- **[L-D02] CI/CD 통합**: PR 코멘트에 트래픽 추이 차트(ASCII 또는 Markdown) 포함 시도.
