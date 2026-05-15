# 🐳 Docker: 00. 클러스터 구축 (Cluster Setup)

Docker Compose를 사용하여 MigraGuard 전체 에코시스템(DB, 에이전트, 부하 생성기)을 배포합니다.

---

## 1. 초기 배포
타겟 데이터베이스, 백그라운드 수집기, 부하 생성기를 빌드하고 시작합니다.
```bash
docker compose up -d --build
```

## 2. 서비스 개요
- **`db`**: PostgreSQL 15 인스턴스.
- **`agent`**: 백그라운드 수집기.
- **`load-generator`**: 실시간 트래픽 시뮬레이터.
- **`analyze`**: 리스크 평가용 CLI 도구.

## 3. 영구 저장소 (Persistence)
- SQLite 데이터는 `migraguard-data` 볼륨에 저장됩니다.
- PostgreSQL 데이터는 `postgres-data` 볼륨에 저장됩니다.

---

## 💡 개발자 팁: 소스 코드 변경 시 이미지 갱신
Go 소스 코드(예: 새로운 명령어 추가)를 수정한 후에는 **반드시 이미지를 다시 빌드**해야 컨테이너 내부에 반영됩니다. 이제 `analyze` 서비스도 기본 빌드 대상에 포함되어 있으므로, 평소처럼 아래 명령어를 실행하면 한꺼번에 최신화됩니다.

```bash
# 모든 서비스(Analyze 포함)를 최신 코드로 빌드 및 실행
docker compose up -d --build
```
