# 🏜️ MigraGuard 시뮬레이션 샌드박스 운용 가이드

본 문서는 실제 PostgreSQL 데이터베이스 연결 없이도 가상의 부하 환경을 구축하고 리스크를 검증하는 **Simulation Sandbox** 기능의 상세 사용법과 수학적 모델을 안내합니다.

---

## 1. 개요 (Overview)

MigraGuard의 시뮬레이션 샌드박스는 연구 및 검증의 **재현성(Reproducibility)**을 위해 설계되었습니다. 7일치 시계열 워크로드 데이터를 수학적 모델을 기반으로 즉시 생성하여 오프라인 실험 환경을 제공합니다.

### 핵심 가치
*   **오프라인 검증**: DB 연결 없이도 리스크 엔진의 모든 시나리오 테스트 가능.
*   **재현성 확보**: 동일한 YAML 시나리오를 통해 전 세계 어디서든 동일한 실험 결과 도출.
*   **극단적 상황 모사**: Flash Sale, 시스템 장애 등 실제 환경에서 재현하기 어려운 케이스 자유 설계.

---

## 2. 시나리오 정의 (YAML Schema)

샌드박스 실험의 설계도인 YAML 파일의 주요 필드 설명입니다.

| 필드 | 설명 | 비고 |
| :--- | :--- | :--- |
| `target_table` | 분석의 주 대상이 될 테이블명 | 필수 |
| `pg_state` | 분석 시점의 Mock 운영 상태 (Size, P99 등) | 필수 |
| `base_tps` / `peak_tps` | 시간대별 기준 트래픽 부하량 | 필수 |
| `weekly_pattern` | 주말 트래픽 40% 감소 패턴 적용 여부 | Boolean |
| `events` | 특정 시점의 돌발 트래픽(Flash Sale 등) 정의 | Array |

---

## 3. 수학적 워크로드 모델 (Mathematical Models)

샌드박스 엔진은 현실적인 데이터 생성을 위해 다음 수식들을 사용합니다.

### 3.1. 일간/주간 트래픽 모델
시간 $t$에 따른 TPS($\lambda$)는 다음과 같이 결정됩니다:
$$\lambda(t) = (\lambda_{base} + \text{Sine}(t) \times (\lambda_{peak} - \lambda_{base})) \times M_{week} \times M_{event} + \epsilon$$
*   $\text{Sine}(t)$: 14시와 20시에 정점을 찍는 이중 사인파 모델.
*   $\epsilon$: 가우시안 노이즈 (불규칙성 모사).

### 3.2. 지표 상관관계 모델
*   **지연 시간(P99)**: TPS 상승 시 지수 함수적으로 증가.
    $$P99 = P99_{base} \times e^{(\frac{\lambda}{\lambda_{peak}} - 1)}$$
*   **커넥션 수**: TPS에 비례하여 선형적으로 증가.

---

## 4. 명령어 사용법 (Usage)

### 4.1. 샌드박스 생성
```bash
migraguard simulate --scenario scenario.yaml
```

### 4.2. 오프라인 리스크 분석
```bash
migraguard analyze migration.sql --sandbox scenario.db
```

---

## 5. 활용 사례
- CI/CD 파이프라인의 사전 리스크 게이트키핑 테스트.
- TPS 및 테이블 크기 변화에 따른 리스크 점수 추이 연구.
- 데이터베이스 서버 구축 없는 교육 및 데모 환경 구축.
