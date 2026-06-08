# 🖥️ 네이티브: 01. 에이전트 (Agent / Collector)

에이전트는 백그라운드에서 PostgreSQL 데이터베이스를 폴링하여 과거 워크로드 프로필을 구축하는 메트릭 수집기입니다.

---

## 1. 실행 예시

**Bash (Unix/macOS/Linux)**
```bash
./migraguard agent --db "postgres://user:pass@localhost:5432/db"
```

**PowerShell (Windows)**
```powershell
.\migraguard.exe agent --db "postgres://user:pass@localhost:5432/db"
```

**CMD (Windows)**
```cmd
migraguard.exe agent --db "postgres://user:pass@localhost:5432/db"
```

---

## 2. 상세 플래그 참조

| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--db` | - | (설정 참조) | **대상 PostgreSQL 연결 문자열 (DSN)**. 모니터링할 데이터베이스를 지정합니다. `pg_stat_statements`에 대한 접근 권한이 필요합니다. |
| `--sqlite` | - | `migraguard.db` | **로컬 저장소 경로**. 워크로드 이력이 저장되는 SQLite 파일입니다. 이 파일은 나중에 `analyze` 명령에서 사용됩니다. |
| `--interval` | - | `60` | **폴링 주기 (초)**. 에이전트가 메트릭을 캡처하는 빈도를 정의합니다. 값이 작을수록 정밀도가 높아지지만 모니터링 오버헤드가 증가합니다. |
| `--retention` | - | `7` | **데이터 보관 기간 (일)**. 에이전트는 디스크 공간 관리를 위해 이 값보다 오래된 메트릭을 자동으로 삭제합니다. |
| `--tables` | - | - | **테이블별 동적 지표 수집 대상**. 쉼표로 구분한 테이블 목록(예: `orders,users`)입니다. 생략하면 `pg_stat_statements` 기반 workload snapshot은 수집하지만, `table_metrics`에 저장되는 테이블별 크기/TPS/P99/connection 지표는 별도로 기록하지 않습니다. |

---

## 3. 권장 사항
*   **운영 환경**: 지속적인 데이터 수집을 위해 에이전트를 시스템 서비스(`systemd` 또는 Windows 서비스)로 실행하십시오.
*   **정밀도**: 부하가 높은 시스템의 경우, 더 정밀한 리스크 평가를 위해 주기를 `30`으로 설정하는 것을 권장합니다.
*   **분석 대상 테이블**: 이후 `analyze --forecast` 또는 테이블별 baseline 분석을 사용할 계획이라면, 분석할 테이블을 `--tables`에 명시하는 것을 권장합니다.
