# 🐳 Docker: 04. 샌드박스 내보내기 (Export Sandbox)

기존 SQLite 샌드박스 데이터베이스에서 메트릭을 CSV로 내보냅니다.

---

## 1. 실행 명령어 (통합)
명령어는 모든 쉘에서 동일합니다.
```bash
docker compose run --rm analyze export-sandbox \
  --input ./code/experiments/data/01_steady_normal.db \
  --output ./code/output.csv
```

---

## 2. 파라미터 참조
- `--input`: 원본 SQLite 파일 경로 (`./code/`로 시작).
- `--output`: 대상 CSV 파일 경로 (`./code/`로 시작).
