# 🖥️ 네이티브: 05. 자가 진단 (Check)

Check 명령은 MigraGuard가 사용할 PostgreSQL 연결과 SQLite 저장소를 초기화할 수 있는지 빠르게 확인합니다.

---

## 1. 실행 예시

**Bash**
```bash
./migraguard check --db "postgres://user:pass@host:5432/db"
```

**PowerShell**
```powershell
.\migraguard.exe check --db "postgres://user:pass@host:5432/db"
```

**CMD**
```cmd
migraguard.exe check --db "postgres://user:pass@host:5432/db"
```

---

## 2. 상세 플래그 참조

| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--db` | - | (설정 참조) | **대상 PostgreSQL URL**. PostgreSQL 연결과 ping 가능 여부를 확인합니다. |
| `--sqlite` | - | (설정 참조) | **대상 SQLite 경로**. MigraGuard SQLite 저장소를 열고 필요한 스키마를 초기화할 수 있는지 확인합니다. |

---

## 3. 현재 확인 범위
1.  **PostgreSQL 연결**: DSN 파싱, connection pool 생성, ping 성공 여부를 확인합니다.
2.  **SQLite 초기화**: 지정된 SQLite 경로를 열고 MigraGuard가 사용하는 기본 테이블을 생성할 수 있는지 확인합니다.

주의: 현재 `check` 명령은 `pg_stat_statements` 확장 로드 여부, `shared_preload_libraries` 설정, 디스크 여유 공간을 별도 쿼리로 진단하지 않습니다. 이러한 항목은 `agent` 실행 또는 운영 DB 설정 점검 단계에서 별도로 확인해야 합니다.
