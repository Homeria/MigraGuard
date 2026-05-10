# 🖥️ 네이티브: 05. 자가 진단 (Check)

Check 명령은 MigraGuard 환경 및 연결 상태에 대한 자체 진단을 수행합니다.

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
| `--db` | - | (설정 참조) | **대상 PostgreSQL URL**. 연결을 확인하고 `pg_stat_statements` 확장이 로드되었는지 점검합니다. |
| `--sqlite` | - | (설정 참조) | **대상 SQLite 경로**. 파일이 존재하고 현재 사용자가 쓰기 권한을 가지고 있는지 확인합니다. |

---

## 3. 진단 체크리스트
1.  **PostgreSQL 연결**: DNS 확인 및 인증 상태를 점검합니다.
2.  **확장 도구 검증**: `pg_stat_statements`가 `shared_preload_libraries`에 포함되어 있는지 확인합니다.
3.  **SQLite 건전성**: 디스크 공간 및 쓰기 권한을 확인합니다.
