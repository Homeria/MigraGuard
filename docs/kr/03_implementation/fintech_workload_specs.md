# ⚙️ 핀테크 워크로드 구현 명세

본 문서는 연구 검증을 위해 v3.8에서 구현된 고동시성 트랜잭션 모델 및 테이블 관계를 상세히 설명합니다.

---

## 1. 트랜잭션 흐름 (복합 주문 시뮬레이션)

`load_generator`는 단순한 개별 테이블 삽입이 아닌, 실제 환경의 핀테크 주문 트랜잭션을 시뮬레이션합니다. 모든 "주문 성공" 지표는 다음 ACID 트랜잭션을 포함합니다.

1.  **재고 확인 및 차감 (`inventory_stocks`):**
    - `UPDATE inventory_stocks SET stock_quantity = stock_quantity - 1 WHERE product_id = ? AND stock_quantity >= 1`
    - *목적:* 인기 상품 아이템에 대한 행 수준(Row-level) 락 경합 재현.
2.  **잔액 확인 및 차감 (`account_balances`):**
    - `UPDATE account_balances SET balance = balance - price WHERE user_id = ? AND balance >= price`
    - *목적:* 개별 사용자 레코드에 대한 고빈도 업데이트 상황 모사.
3.  **주문 생성 및 영속화 (`orders`):**
    - `INSERT INTO orders (...) RETURNING id`
    - *목적:* 메인 트랜잭션 테이블의 규모를 키워 테이블 재작성(Rewrite) 비용 증가 유도.
4.  **감사 로그 기록 (`order_event_logs`):**
    - `INSERT INTO order_event_logs (...)`
    - *목적:* 대규모 Append-only 로그 트래픽 및 I/O 부하 모사.

---

## 2. 연구용 테이블 특성

### A. account_balances (핫스팟)
- **특징:** 레코드 크기는 작으나 업데이트 빈도가 극도로 높음.
- **리스크 시나리오:** 이 테이블에 표준 인덱스를 추가하거나 컬럼 타입을 변경할 경우, 모든 금융 트랜잭션이 차단되어 즉각적인 시스템 마비 유발.

### B. orders (헤비 바디)
- **특징:** 데이터 부피가 크고 복잡한 인덱스 보유.
- **리스크 시나리오:** 테이블 재작성(Rewrite)을 유발하는 작업 시 MigraGuard가 정확한 $T_{ddl}$ 값을 예측하는지 확인.

### C. inventory_stocks (경합점)
- **특징:** 엄격한 무결성 제약 조건 보유.
- **리스크 시나리오:** "Flash Sale" 상황에서 병목 지점이 됨. 제약 조건 변경 시 커넥션 풀 대기열 폭증 현상 분석.
