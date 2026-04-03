# 🛡️ MigraGuard v3.0 상세 로직 명세서 (Logic Specification)

본 문서는 MigraGuard v3.0의 각 로직에 고유 번호(Logic ID)를 부여하여 관리합니다.

---

## 🚀 Phase 0: CLI 진입점 및 오케스트레이션
**파일:** `cmd/migraguard/analyze.go`

- **[L01] SQL 로드**: 마이그레이션 SQL 파일을 텍스트로 읽어들임.
- **[L02] 정적 분석 호출**: `parser.ParseSQL()`을 통한 DDL 특성 파악.
- **[L03] DB 동적 지표 수집**: `db.PostgresAdapter`를 통한 실시간 부하 데이터 획득.
- **[L04] 리스크 계산**: `engine.AnalyzeRisk()`를 통한 정량적 위험도 산출.
- **[L05] 최종 판단**: 산출된 결과를 바탕으로 배포 승인/거부 결정.

---

## 🔍 Phase 1: 정적 분석 (Static AST Analysis)
**파일:** `internal/parser/ast.go`

- **[L11] AST 변환**: SQL 문장을 PostgreSQL 파서를 통해 트리 구조로 변환.
- **[L12] 작업 식별**: DDL 유형 및 테이블 재작성($F_{rewrite}$) 필요성 판단.

---

## 📊 Phase 2: 실시간 동적 지표 수집
**파일:** `internal/db/collector.go`

- **[L21] 테이블 크기 ($S_{table}$)**: 대상 테이블의 물리적 용량 측정.
- **[L22] 트래픽 처리량 ($\lambda$)**: 시스템의 초당 트랜잭션 수(TPS) 측정.
- **[L23] 활성 커넥션 ($C_{active}$)**: 현재 DB를 점유 중인 세션 수 측정.
- **[L24] 복제 지연 ($Lag_{repl}$)**: 마스터-슬레이브 간 동기화 지연 시간 측정.

---

## 🧠 Phase 3: 리스크 엔진 (v3.0 Queuing Model)
**파일:** `internal/engine/risk.go`

- **[L31] DDL 시간 추정 ($T_{ddl}$)**: $S_{table} / Disk\_IO$ 기반 물리 작업 시간 예측.
- **[L32] 블로킹 시간 산출 ($T_{block}$)**: $T_{p99} + T_{ddl} + Lag_{repl}$ 기반 서비스 영향 시간 예측.
- **[L33] 큐 스파이크 산출 ($C_{peak}$)**: $C_{active} + (\lambda \times T_{block})$ 기반 예상 커넥션 부하 예측.
- **[L34] 회복 시간 산출 ($T_{rec}$)**: $(C_{peak} - C_{max}) / (\mu_{max} - \lambda)$ 기반 정상화 시간 예측.
- **[L35] 위험도 점수 산출**: $(C_{peak} / C_{max}) \times 100$ 기반 최종 점수 산출.

---

## 🚦 Phase 4: 리포팅 및 게이트키핑
**파일:** `cmd/migraguard/analyze.go` & `internal/reporter/console.go`

- **[L41] 결과 리포팅**: 산출된 수치를 시각화하여 터미널에 출력.
- **[L42] 프로세스 제어**: 위험 수준에 따라 Exit Code 0 또는 1 반환.
