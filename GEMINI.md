# 🛡️ MigraGuard 프로젝트 진행 상황 (Session Handover)

본 문서는 다른 세션에서 작업을 이어받기 위한 가이드 및 현재 상태 기록입니다.

## 📅 마지막 업데이트: 2026-03-31
- **현재 브랜치:** `cli-core`
- **주요 목표:** DB 마이그레이션 시 락 경합 방지를 위한 CLI 게이트키퍼 구축

## ✅ 지금까지 완료된 작업
1. **요구사항 분석 및 설계 문서화:**
   - `README.md` 및 `spec/total_spec.pdf` 분석 완료.
   - `spec/project_summary.md`: 프로젝트 개요, 핵심 가치, 전체 기능 명세 정리.
   - `spec/tech_spec.md`: 시스템 아키텍처, 데이터 흐름, Risk Score 산출 알고리즘 상세 설계.

2. **CLI 코어 보일러플레이트 구현:**
   - `spf13/cobra` 프레임워크 기반 CLI 구조 구축.
   - `cmd/migraguard/main.go`, `root.go`, `analyze.go`, `check.go` 작성 완료.
   - `analyze`: SQL 파일 정적 분석 및 위험도 평가 명령어 (뼈대).
   - `check`: 배포 직전 `pg_stat_activity` 점검 명령어 (뼈대).
   - 빌드 및 실행 확인 완료 (`go build -o migraguard ./cmd/migraguard`).

3. **환경 설정:**
   - 의존성 관리 (`go.mod`, `go.sum`) 업데이트 완료.
   - 불필요한 파일 정리 (루트 `main.go` 삭제 등).

## 🚀 다음 세션에서 수행할 작업 (Next Steps)
1. **`internal/parser` 구현:**
   - `pganalyze/pg_query_go`를 활용하여 `internal/parser/ast.go` 완성.
   - DDL에서 타겟 테이블 및 요구 락 레벨(AccessExclusiveLock 등) 추출 로직 구현.
2. **`internal/db` 구현:**
   - `jackc/pgx`를 사용하여 PostgreSQL 연결부 작성.
   - `pg_stat_statements` 및 `pg_stat_activity` 조회 쿼리 구현.
3. **`internal/engine` 구현:**
   - `spec/tech_spec.md`에 정의된 Risk Score 산출 알고리즘 코딩.
4. **시계열 데이터 저장 (SQLite):**
   - 워크로드 스냅샷 저장을 위한 SQLite 스키마 설계 및 연동.

---
**세션 연결 가이드:** 다음 세션 시작 시 "GEMINI.md 파일을 읽고 `cli-core` 브랜치 작업을 이어서 진행해줘"라고 명령하세요.
