# 🧩 MigraGuard SQL 파서 로직 명세

본 문서는 MigraGuard가 DDL SQL을 분석하여 테이블 재작성(Table Rewrite) 및 락 레벨을 판별하는 기술적 메커니즘을 설명합니다.

---

## 1. 개요

MigraGuard는 `pg_query_go`를 통해 PostgreSQL의 공식 C-파서를 활용합니다. 이는 운영 환경에서 실행될 SQL의 구문 해석에 있어 100%의 정확도를 보장합니다.

---

## 2. 분석 파이프라인

1.  **AST 생성**: SQL 문자열을 Abstract Syntax Tree로 변환합니다.
2.  **DDL 식별**: `AlterTableStmt`, `CreateStmt`, `IndexStmt` 등 주요 DDL 노드를 추출합니다.
3.  **위험 지표 판별**:
    *   **Lock Level**: PostgreSQL 내부 락 매트릭스를 기반으로 등급(1~8)을 할당합니다.
    *   **Rewrite Required**: 컬럼 타입 변경, 제약 조건 추가 등 테이블 전체 스캔 및 재작성이 필요한지 확인합니다.
    *   **Metadata Only**: 단순 기본값 설정이나 NULL 허용 변경 등 카탈로그 업데이트만으로 끝나는지 판별합니다.

---

## 3. 주요 판별 로직 예시

| SQL 문구 | 판별 결과 | 이유 |
| :--- | :--- | :--- |
| `ALTER TABLE orders ADD COLUMN age int;` | **Safe (Metadata)** | 단순 컬럼 추가는 메타데이터 갱신으로 종료. |
| `ALTER TABLE orders ALTER COLUMN no TYPE bigint;` | **Danger (Rewrite)** | 타입 변경은 데이터 전체 재작성을 유발 (Storage I/O 발생). |
| `CREATE INDEX idx_name ON users(name);` | **Warning (Lock)** | 표준 인덱스 생성은 `ShareLock`을 획득하여 쓰기 차단. |

---

## 4. 아키텍처적 장점
- **신뢰성**: PostgreSQL 공식 파서 라이브러리 사용으로 문법 오류 탐지 정확도 극대화.
- **예측 가능성**: 정적 분석 단계에서 이미 락의 성격과 실행 시간의 기저(Base)를 확정함.
