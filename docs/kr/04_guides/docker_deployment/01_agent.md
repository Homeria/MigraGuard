# 🐳 Docker: 01. 에이전트 (Agent / Collector)

에이전트 서비스는 백그라운드에서 워크로드 패턴을 수집합니다.

---

## 1. 실행 명령어 (통합)
명령어는 모든 쉘(Bash, PowerShell, CMD)에서 동일합니다.
```bash
docker compose logs -f agent
```

## 2. 설정 변경
에이전트 설정을 바꾸려면 `docker-compose.yml`의 `command`를 수정하고 재시작하십시오.
```bash
docker compose up -d agent
```

---

## 3. 주요 플래그
| 플래그 | 기본값 | 설명 |
| :--- | :--- | :--- |
| `--db` | (설정 참조) | 대상 DB. 서비스 이름인 `db`를 사용합니다. |
| `--sqlite` | `/app/data/migraguard.db` | 내부 볼륨 내 저장 경로. |
| `--interval` | `60` | 수집 주기 (초). |
