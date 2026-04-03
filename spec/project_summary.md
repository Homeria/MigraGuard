# MigraGuard 프로젝트 개요 및 분석 (v3.1 에이전트 모델)

## 1. 프로젝트 정의 (What is MigraGuard?)
**MigraGuard**는 데이터베이스 스키마 변경(DDL) 시 발생할 수 있는 락(Lock) 경합 및 서비스 장애를 사전에 차단하는 **DevSecOps 에이전트 및 CLI 도구**입니다. 운영 DB의 트래픽 패턴을 **상시 수집(Agent)**하여 시계열 데이터베이스(SQLite)를 구축하고, 마이그레이션 스크립트를 분석할 때 이 데이터를 즉각 활용하여 배포 안전성을 정량적으로 평가합니다.

## 2. 핵심 가치 (Core Value)
> "트래픽을 아는 에이전트가 DDL 시한폭탄을 해체한다."
- **상시 워크로드 감시 (Continuous Monitoring):** 분석 시점의 데이터뿐만 아니라, 과거의 트래픽 패턴을 학습하여 리스크를 판단함.
- **즉각적인 분석 피드백 (Instant Feedback):** 이미 수집된 데이터를 사용하므로 분석 시 대기 시간이 없음.
- **배포 골든타임 추천 (Safe Window):** 축적된 데이터를 기반으로 하루 중 트래픽이 가장 낮은 최적의 배포 시간대를 제시함.
- **자동화된 게이트키퍼:** CI/CD 파이프라인과 통합되어 위험한 배포를 원천 차단.

## 3. 주요 기능 명세
1. **Agent 모드 (Background Collector):**
   - 운영 DB 옆에서 상시 가동하며 `pg_stat_statements` 스냅샷을 1분/1초 단위로 SQLite에 기록.
   - 오래된 데이터를 자동으로 정리하는 데이터 보존 정책(Retention Policy) 수행.
2. **Analyze 모드 (Foreground CLI):**
   - SQL AST 파싱을 통해 DDL 영향도 파악.
   - Agent가 쌓아둔 SQLite 데이터를 쿼리하여 즉각적인 리스크 분석 및 리포팅.
3. **위험도 산출 모델 (v3.1):**
   - 현재 트래픽($\lambda$)과 과거 평균/최대 트래픽을 비교하여 이상 징후 감지.
   - 큐잉 모델(Queuing Model)을 통한 커넥션 고갈 위험도 점수화.
4. **시각화 및 리포팅:** CLI 출력 및 GitHub PR 마크다운 코멘트 자동 작성.

## 4. 기술 스택
- **언어:** Go (Golang)
- **데이터베이스:** SQLite (시계열 지표 저장소), PostgreSQL (대상 운영 DB)
- **인프라:** Docker (Agent와 CLI 간 볼륨 공유를 통한 데이터 연동)
