# 🖥️ 네이티브: 06. 부하 생성기 (LoadGen)

LoadGen은 실제 PostgreSQL 데이터베이스에 현실적인 핀테크 트랜잭션을 발생시켜, 압박 상황에서의 모니터링 및 리스크 분석을 테스트하기 위한 독립 바이너리입니다.

---

## 1. 실행 예시

**Bash**
```bash
./loadgen --db "postgres://user:pass@host:5432/db" --conns 20 --profile steady
```

**PowerShell**
```powershell
.\loadgen.exe --db "postgres://user:pass@host:5432/db" --conns 20 --profile steady
```

**CMD**
```cmd
loadgen.exe --db "postgres://user:pass@host:5432/db" --conns 20 --profile steady
```

---

## 2. 상세 플래그 참조

| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--db` | - | **(필수)** | **대상 PostgreSQL URL**. 부하를 주입할 데이터베이스의 DSN입니다. |
| `--conns` | - | `10` | **동시성 (워커 수)**. 트랜잭션을 실행하는 동시 스레드 수입니다. 값이 높을수록 잠금 경합이 심해집니다. |
| `--profile` | - | `steady` | **부하 프로파일**. 트랜잭션 조합을 정의합니다: <br> - `steady`: 읽기/쓰기 70:30 비율. <br> - `flash-sale`: 쓰기 집중 30:70 비율. <br> - `read-heavy`: 읽기 전용 검색 95:5 비율. |

---

## 3. 트랜잭션 상세 정보
*   **Browse**: 사용자, 상품 및 재고 정보를 `SELECT`합니다.
*   **Order**: `inventory_stocks` 업데이트, `account_balances` 차감, `orders` 삽입을 포함하는 전체 트랜잭션입니다.
