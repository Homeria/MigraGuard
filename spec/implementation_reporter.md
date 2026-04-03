# MigraGuard Phase 4: Reporter & Gatekeeper 구현 리포트 (v3.0 고도화)

## 1. 구현 개요
- **목적**: 분석 결과의 가독성 높은 리포팅 및 고위험 배포 시 CI/CD 파이프라인 자동 차단.
- **핵심 모듈**: `cmd/migraguard/analyze.go`

## 2. 주요 기능 및 로직 상세
- **통합 분석 파이프라인**:
    - `ParseSQL` -> `PostgresAdapter.GetTableDynamicMetrics` -> `RiskEngine.AnalyzeRisk` 순차 실행.
- **정량적 리포팅**:
    - v3.0 모델의 핵심 지표($RiskScore$, $T_{ddl}$, $T_{block}$, $C_{peak}$, $T_{rec}$)를 터미널에 시각적으로 출력.
    - 리스크 레벨(`Danger`, `Warning`, `Safe`)에 따른 직관적 피드백 제공.
- **게이트키핑 (Gatekeeping)**:
    - 리스크 분석 결과 중 하나라도 `Danger`일 경우 `os.Exit(1)`을 호출하여 CI/CD 프로세스 중단.
- **환경 유연성**:
    - `DATABASE_URL`, `SQLITE_PATH` 환경 변수를 지원하여 다양한 런타임 환경에 대응.

## 3. 리포트 출력 예시
```text
--- 🛡️ MigraGuard v3.0 Risk Analysis Report ---
[Target Table: users | Operation: ALTER]
  - Risk Level:       [Danger]
  - Risk Score:       94.50%
  - Estimated DDL:    1200.00 ms
  - Blocking Time:    1450.00 ms
  - Peak Connections: 945 conns
  - Recovery Time:    2500.00 ms
  🚨 CRITICAL: Permanent system failure predicted!
```

## 4. 향후 과제
- **GitHub PR 코멘터**: CI 환경에서 분석 리포트를 Markdown 포맷으로 변환하여 PR에 자동 게시하는 기능 추가.
- **설정 파일 지원**: `migraguard.yaml` 파일을 통한 임계치(Threshold) 및 상수 커스터마이징 기능 구현.
