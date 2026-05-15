# 🐳 Docker: 06. 부하 생성기 (LoadGen)

`db` 컨테이너에 실제 핀테크 트래픽을 시뮬레이션합니다.

---

## 1. 실행 명령어 (통합)
명령어는 모든 쉘에서 동일합니다.
```bash
docker compose run -d --name stress-test load-generator --conns 50 --profile flash-sale
```

---

## 2. 주요 플래그
| 플래그 | 기본값 | 설명 |
| :--- | :--- | :--- |
| `--conns` | `10` | 워커 동시성 수. |
| `--profile` | `steady` | `steady`, `flash-sale`, `read-heavy`. |
