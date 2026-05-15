# 🐳 Docker: 05. 자가 진단 (Check)

연결성 및 상태를 검증합니다.

---

## 1. 실행 명령어 (통합)
명령어는 모든 쉘에서 동일합니다.
```bash
docker compose run --rm analyze check
```
---
- `db` 컨테이너 연결 상태 확인.
- 공유 `migraguard.db` 쓰기 권한 확인.
