# 🎓 캡스톤 디자인 최종 보고서: MigraGuard
**부제: 실시간 트래픽 데이터 기반의 지능형 데이터베이스 마이그레이션 서킷 브레이커(Circuit Breaker) 및 예측 시스템**

---

## 1. 서론 (Introduction)

### 1.1. 배경 및 필요성
현대의 소프트웨어 개발 환경(CI/CD)에서는 무중단 배포(Zero-Downtime Deployment)가 필수적입니다. 그러나 애플리케이션 코드는 무중단 배포가 용이한 반면, 데이터베이스 스키마 변경(DDL)은 여전히 가장 큰 위험 요소로 남아있습니다. 
기존의 SQL Linter(예: sqlfluff) 도구들은 쿼리의 문법이나 정적인 안티패턴만을 검사할 뿐, **"해당 쿼리가 현재 운영 중인 서버의 트래픽 환경에서 얼마나 위험한지"**를 판단하지 못합니다. 트래픽이 적은 새벽에는 안전한 DDL 작업이 트래픽이 몰리는 점심시간에는 심각한 락 경합(Lock Contention)과 서비스 전면 장애를 유발할 수 있습니다.

### 1.2. 프로젝트 목표
본 프로젝트(MigraGuard)의 목표는 **정적인 인프라 환경 변수(Config)**와 **동적인 실시간 워크로드 데이터(TPS, P99 등)**를 결합하여 DDL의 실행 위험도를 정량적으로 평가하고, 자동화된 파이프라인에서 위험을 차단(Circuit Breaker)하는 DevSecOps 도구를 개발하는 것입니다. 더 나아가, 향후 24시간의 트래픽을 예측하여 가장 안전하게 작업할 수 있는 **'골든 윈도우(Golden Window)'**를 시각적 히트맵과 함께 제안하는 의사결정 지원 시스템을 구축합니다.

---

## 2. 시스템 아키텍처 (System Architecture)

MigraGuard는 Go 언어 기반의 **SDK-First Architecture**를 채택하였으며, CLI 툴은 Cobra 프레임워크를 통해 구현되었습니다.

### 2.1. 주요 컴포넌트
1.  **Core SDK (`pkg/migraguard`)**: 핵심 비즈니스 로직(리스크 평가, 메트릭 수집, 시뮬레이션)을 독립적인 라이브러리 형태로 캡슐화하여 높은 재사용성을 보장합니다.
2.  **CLI Interface (`cmd/migraguard`)**: 개발자 및 CI/CD 파이프라인(GitHub Actions 등)에서 즉시 사용할 수 있는 단일 바이너리(Executable)를 제공합니다. (`analyze`, `simulate`, `check` 등의 명령어 지원)
3.  **Agent & SQLite Storage**: 대상 PostgreSQL에서 실시간 TPS, P99 지연 시간, 커넥션 수 등의 메트릭을 주기적으로 수집하여 로컬 SQLite(`migraguard.db`)에 경량 저장합니다.
4.  **Simulation Sandbox (`VirtualPGAdapter`)**: 실제 DB에 부하를 주지 않고 YAML 기반의 선언적 시나리오를 통해 가상의 환경을 구축하여 DDL 리스크를 사전 검증합니다.

---

## 3. 핵심 알고리즘: 5단계 정량적 리스크 모델 (5-Step Risk Model)

MigraGuard의 차별점은 단순 룰-베이스 검사가 아닌, 대기열 이론(Queueing Theory)과 Little's Law에 근거한 수학적 모델링입니다. 전략 패턴(Strategy Pattern)으로 구현된 5개의 Evaluator가 순차적으로 위험을 계산합니다.

*   **정적 상수 (Config)**: $\mu_{max}$ (최대 처리량), $DiskIO$, $C_{max}$ (최대 커넥션)
*   **동적 변수 (Metric)**: $TPS$, $P99$, $Lag$ (복제 지연)

### Step 1: DDL 소요 시간 예측 ($T_{ddl}$)
*   단순 메타데이터 변경은 기본 시간($T_{meta}$)을 할당.
*   Table Rewrite가 필요한 경우(예: 타입 변경) 테이블 사이즈와 디스크 I/O를 비례하여 계산: $T_{ddl} = (TableSize / DiskIO) \times 1000$

### Step 2: 서비스 차단 시간 계산 ($T_{block}$)
*   DDL 실행 대기 및 락 획득 과정에서 발생하는 총 블로킹 시간:
*   $T_{block} = (P99 + T_{ddl} + Lag) \times LockImpact$ (락 수준에 따른 가중치 적용)

### Step 3: 최대 연결 수 폭증 예측 ($C_{peak}$)
*   차단된 시간 동안 해소되지 못하고 쌓이는 애플리케이션 커넥션 수:
*   $C_{peak} = ActiveConns + (\lambda_{tps} \times T_{block})$

### Step 4: 시스템 회복 시간 평가 ($T_{rec}$)
*   락 해제 후 밀린 큐(Backlog)를 소화하는 데 걸리는 시간. 현재 TPS가 한계치($\mu_{max}$)를 초과하면 영구 장애(Permanent Failure)로 간주.
*   $T_{rec} = (C_{peak} - C_{max}) / (\mu_{max} - TPS)$

### Step 5: 리스크 점수 산출 및 차단 (Risk Decision)
*   최대 수용 커넥션($C_{max}$) 대비 $C_{peak}$의 비율을 점수화: $(C_{peak} / C_{max}) \times 100$
*   이 점수를 기반으로 **Safe / Warning / Danger** 레벨을 판정하여 파이프라인 차단 여부를 결정.

---

## 4. 다차원 시뮬레이션 및 평가 (Evaluation & Batch Analysis)

MigraGuard의 모델 검증을 위해 대규모 **다차원 배치 시뮬레이션(Multi-Dimensional Batch Analysis)**을 수행하였습니다.

### 4.1. 실험 설계
*   **시나리오 (10종)**: 안정적 워크로드, 사인 곡선(일간/주간 주기), 플래시 세일(스파이크), 연결 고갈 등 다양한 트래픽 환경.
*   **테스트 케이스 (20종)**: 안전한 DDL(인덱스 동시 생성 등)부터 위험한 DDL(테이블 리라이트, 컬럼 삭제 등).
*   **시스템 설정 (3종)**: Default, LowCapacity($\mu_{max}=2000$), HighCapacity($\mu_{max}=10000$).
*   총 600개의 조합에 대한 검증을 스크립트로 자동화하여 실행.

### 4.2. 주요 결과 및 성능 (Tie-Breaker 도입)
1.  **환경 적응성 증명**: 동일한 '컬럼 타입 변경(Rewrite)' DDL이 HighCapacity 환경에서는 1600%대의 리스크를 보인 반면, LowCapacity에서는 6600%로 폭증하여 **정적 설정과 동적 트래픽의 유기적 결합**이 완벽히 동작함을 증명했습니다.
2.  **골든 윈도우 정확도 개선**: 초기에는 점수가 완전히 동일하게 낮은 '안전 구간(Safe Zone)'들이 연속될 때 단순히 가장 이른 시간(00:00)을 제안하는 한계가 있었습니다. 이를 해결하기 위해 리스크 점수가 동일할 경우 **최저 TPS 구간을 우선하는 Tie-Breaker 알고리즘**을 적용하여, 물리적 트래픽이 가장 바닥을 찍는 "진정한 최적의 시간(예: 02:00)"을 찾아내도록 개선했습니다.
3.  **예측 시각화 (Dual Y-Axis Heatmap)**: TPS 추이와 예측된 P99 지연 시간을 꺾은선 그래프(이중 Y축)로 표시하고, 배경색으로 리스크 레벨(Safe/Warning/Danger)을 직관적으로 보여주는 히트맵을 생성하여 뛰어난 의사결정 UX를 제공합니다.

---

## 5. 결론 및 향후 과제 (Conclusion & Future Work)

MigraGuard는 단순한 SQL 검사기를 넘어서, 시스템의 물리적 한계치와 실시간 트래픽 데이터를 바탕으로 **"언제, 어떻게 배포해야 하는가"**를 알려주는 지능형 DevSecOps 도구로 완성되었습니다. CI/CD 파이프라인에 이 툴을 연동함으로써 기업은 마이그레이션으로 인한 서비스 장애 확률을 극적으로 낮출 수 있습니다.

**향후 과제 (Future Work)**
*   **Adaptive Recommendation (머신러닝 도입)**: 현재는 사용자가 yaml을 통해 $\mu_{max}$ 등을 수동으로 설정해야 하지만, 향후 수집된 메트릭 데이터를 기반으로 시스템의 한계치를 자동으로 역산하여 추천하는 기능을 도입할 예정입니다.

---
<br>

# 📊 [부록] 캡스톤 디자인 발표용 PPT 구성 기획 (12 Slides)

**[Slide 1] 타이틀 슬라이드**
*   **제목**: MigraGuard - 실시간 트래픽 기반 지능형 DB 마이그레이션 서킷 브레이커
*   **발표자/팀명**
*   **핵심 비주얼**: CI/CD 파이프라인과 트래픽 히트맵이 교차하는 다이어그램

**[Slide 2] Problem Statement (문제 제기)**
*   무중단 배포 시대, 애플리케이션 롤백은 쉽지만 DB 스키마(DDL) 롤백은 치명적임.
*   기존 SQL Linter의 한계: "트래픽이 없는 로컬에서는 성공, 트래픽이 몰리는 운영에서는 락 경합으로 인한 전면 장애 발생". 정적 분석의 맹점 지적.

**[Slide 3] Solution: MigraGuard 소개**
*   **핵심 컨셉**: "상황 인지형(Context-Aware) 게이트키퍼"
*   단순 문법 검사가 아닌, 실시간 워크로드(동적)와 시스템 용량(정적)을 융합하여 실제 장애 발생 확률을 계산하고 위험 시 파이프라인을 차단(Circuit Breaking).

**[Slide 4] System Architecture (아키텍처 개요)**
*   SDK-First 디자인 및 Cobra 기반 CLI 툴.
*   Background Agent (메트릭 수집) -> SQLite 로컬 저장소 -> Virtual PG Adapter (시뮬레이션 샌드박스) -> Risk Engine (분석).

**[Slide 5] Core Tech 1: 5-Step 정량적 리스크 모델 (수학적 접근)**
*   대기열 이론(Queueing Theory) 적용.
*   $T_{ddl}$ -> $T_{block}$ (차단 시간) -> $C_{peak}$ (최대 커넥션 폭증) -> $T_{rec}$ (복구 시간).
*   리스크를 추상적인 느낌이 아닌 "수치(%)"로 증명.

**[Slide 6] Core Tech 2: 환경 적응력 (Dynamic vs Static)**
*   인프라 환경(High Capacity vs Low Capacity)에 따라 동일한 DDL도 리스크 점수가 달라짐을 수식과 함께 설명. (개방-폐쇄 원칙을 준수한 Strategy Pattern 언급)

**[Slide 7] Predictive Engine (미래 24시간 예측)**
*   7일간의 과거 데이터를 바탕으로 향후 24시간의 TPS 및 P99 지연 시간 예측.
*   **Tie-Breaker 알고리즘**: 점수가 동일한 안전 구간 내에서도 '가장 물리적 트래픽이 적은' 지점을 찾아내는 골든 윈도우 탐색 로직 소개.

**[Slide 8] Visualization: Predictive Risk Heatmap**
*   생성된 히트맵 이미지 삽입.
*   Dual Y-Axis (TPS & P99) 설명 및 배경색(Safe/Warning/Danger)이 의미하는 바 직관적 해설.

**[Slide 9] Evaluation & Batch Simulation (성능 검증)**
*   다차원 배치 분석 결과 요약 (10 시나리오 x 20 DDL x 3 Configs = 600회 검증 완료).
*   사인 곡선 트래픽이나 플래시 세일(스파이크) 상황에서 시스템이 얼마나 정확하게 Danger Zone을 차단하는지 시연 결과 공유.

**[Slide 10] DevSecOps 파이프라인 통합 (기대 효과)**
*   GitHub Actions 파이프라인 예시 이미지.
*   개발자(Developer)와 DBA/운영자(Ops) 간의 커뮤니케이션 비용 감소 및 배포 안정성 극대화. Blast Radius(장애 반경) 최소화.

**[Slide 11] Conclusion & Future Work (결론 및 향후 계획)**
*   MigraGuard는 단순한 툴이 아닌, 데이터를 기반으로 한 '배포 의사결정 지원 시스템'임.
*   향후 과제: 수집된 메트릭 패턴을 분석하여 최적의 Config 임계값을 자동으로 제안하는 'Adaptive Recommendation(ML 기반)' 고도화 계획.

**[Slide 12] Q&A**
*   질의응답 및 GitHub 레포지토리 링크/QR 코드 제공.
