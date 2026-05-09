# 🛡️ MigraGuard: PostgreSQL 마이그레이션 리스크 게이트키퍼

> **"운영 트래픽에 대한 통찰력이 없는 DDL 배포는 시스템 마비의 시작입니다."**
>
> MigraGuard는 운영 데이터베이스의 실제 트래픽이 마이그레이션 DDL에 미치는 영향을 정량적으로 예측하는 고충실도 리스크 예측 시스템입니다. DDL 시맨틱과 실시간 워크로드 패턴을 교차 분석하여, 보이지 않는 락(Lock) 경합으로 인한 서비스 장애를 사전에 방지합니다.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/Homeria/MigraGuard)](https://goreportcard.com/report/github.com/Homeria/MigraGuard)
[![Research Grade](https://img.shields.io/badge/Architecture-Research--Grade-blueviolet.svg)](#-수학적-리스크-모델)

---

## 📖 목차
- [왜 MigraGuard인가?](#-왜-migraguard인가)
- [시스템 아키텍처](#-시스템-아키텍처)
- [수학적 리스크 모델](#-수학적-리스크-모델)
- [시뮬레이션 샌드박스](#-시뮬레이션-샌드박스)
- [빠른 시작](#-빠른-시작)
- [문서 허브](#-문서-허브)

---

## 🧐 왜 MigraGuard인가?

현대적인 핀테크 및 고트래픽 시스템에서, 단 하나의 `ALTER TABLE` 문은 **배타적 락(Exclusive Lock)**으로 인해 전체 데이터베이스를 마비시킬 수 있습니다. 기존 도구들이 구문 검증에 집중할 때, MigraGuard는 **운영 컨텍스트**에 집중합니다.

- **문제점**: 운영 환경에서의 DDL 실행은 종종 "믿음의 도약"이 됩니다. 개발자는 실제로 실행되기 전까지 작업이 얼마나 걸릴지, 얼마나 많은 커넥션이 블로킹될지 알 수 없습니다.
- **해결책**: MigraGuard는 현재의 TPS, 지연 시간(P99), 복제 지연 지표를 바탕으로 SQL을 분석하여 미래의 리스크를 예측합니다.

---

## 🏗️ 시스템 아키텍처

MigraGuard는 **SDK-First 아키텍처**로 설계되어, 리스크 엔진을 모든 CI/CD 파이프라인이나 모니터링 대시보드에 쉽게 통합할 수 있습니다.

### 이중 프로세스 파이프라인
1.  **MigraGuard Agent**: `pg_stat_statements`에서 메트릭을 수집하고 시계열 워크로드 패턴을 로컬 SQLite 데이터베이스에 저장하는 경량 백그라운드 수집기입니다.
2.  **MigraGuard Analyze**: 마이그레이션 SQL을 AST(추상 구문 트리)로 파싱하고 저장된 패턴을 사용하여 리스크를 평가하는 CLI 도구입니다.

### 고급 디자인 패턴
- **전략 패턴 (Strategy Pattern)**: $T_{ddl}, T_{block}, C_{peak}, T_{rec}$를 위한 모듈화된 리스크 평가 객체.
- **팩토리 패턴 (Factory Pattern)**: 재현 가능한 연구를 위한 **라이브 모드**와 **샌드박스 모드**의 명확한 분리.
- **의존성 주입 (Dependency Injection)**: 높은 테스트 가능성과 관찰 가능성을 위한 완전한 도메인 로직 격리.

---

## 📊 수학적 리스크 모델

MigraGuard의 핵심은 대기 행렬 이론과 데이터베이스 내부 구조에서 도출된 **5단계 정량적 모델**입니다:

1.  **실행 시간 ($T_{ddl}$)**: 테이블 크기와 저장소 처리량을 바탕으로 I/O 비용을 예측합니다.
2.  **블로킹 윈도우 ($T_{block}$)**: $T_{ddl} + P99_{latency} + Lag_{repl}$.
3.  **커넥션 피크 ($C_{peak}$)**: 블로킹 윈도우 동안의 커넥션 폭증을 예측합니다: $C_{active} + (\lambda \times T_{block})$.
4.  **회복 비용 ($T_{rec}$)**: 락 해제 후 적체된 요청을 처리하는 시스템의 능력을 평가합니다.
5.  **리스크 점수화**: 시스템 용량($C_{max}, \mu_{max}$)을 기준으로 **Safe / Warning / Danger** 등급을 할당합니다.

---

## 🏜️ 시뮬레이션 샌드박스 (Research Ready)

학술적 검증 및 "가상 상황(What-if)" 분석을 위해, MigraGuard는 고충실도 **시뮬레이션 엔진**을 제공합니다. 간단한 YAML 시나리오만으로 수개월 치의 현실적인 워크로드 데이터(사인파, 이벤트 스파이크, 가우시안 노이즈)를 생성합니다.

```bash
# 7일간의 시뮬레이션 트래픽 생성
migraguard simulate --scenario experiments/scenarios/03_spike_flash_sale.yaml

# 시뮬레이션된 환경에 대해 분석 실행
migraguard analyze migration.sql --sandbox experiments/data/spike.db
```

---

## 🛠️ 빠른 시작

### 설치
```bash
go install github.com/Homeria/MigraGuard/cmd/migraguard@latest
```

### 1. 모니터링 시작 (에이전트)
```bash
migraguard agent --db "postgres://user:pass@localhost:5432/db"
```

### 2. 마이그레이션 분석 (CLI)
```bash
migraguard analyze ./migrations/001_heavy_alter.sql
```

---

## 📚 문서 허브

상세 문서는 한국어와 영어 모두 제공됩니다.

| 카테고리 | KR (한국어) | EN (English) |
| :--- | :--- | :--- |
| **요구사항** | [KR-01](./docs/kr/01_requirements/functional_spec.md) | [EN-01](./docs/en/01_requirements/functional_spec.md) |
| **아키텍처** | [KR-02](./docs/kr/02_architecture/system_overview.md) | [EN-02](./docs/en/02_architecture/system_overview.md) |
| **구현 상세** | [KR-03](./docs/kr/03_implementation/parser_logic.md) | [EN-03](./docs/en/03_implementation/parser_logic.md) |
| **사용자 가이드** | [KR-04](./docs/kr/04_guides/sandbox_manual.md) | [EN-04](./docs/en/04_guides/sandbox_manual.md) |

---

## 📜 라이선스
**MIT 라이선스**에 따라 배포됩니다. **Homeria / MigraGuard 팀** 제작.
