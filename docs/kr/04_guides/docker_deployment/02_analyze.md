# 🐳 Docker: 02. 리스크 분석 (Analyze)

고정밀 리스크 분석을 실행합니다. **호스트 파일 사용 시 `./code/` 접두사가 필수입니다.**

---

## 💡 필수 사항: 경로 매핑
컨테이너는 호스트의 파일을 `./code/` 디렉토리 하위로 인식합니다.
- **실패**: `analyze experiments/ddl/mig.sql`
- **성공**: `analyze ./code/experiments/ddl/mig.sql`

---

## 1. 실행 명령어 (통합)
기본 실행 명령어는 모든 쉘에서 동일합니다.
```bash
docker compose run --rm analyze analyze ./code/experiments/ddl/001_safe_add_column_orders.sql
```

## 2. 리포트 저장 (쉘별 차이)
결과를 호스트 파일로 저장(리다이렉션)할 때만 쉘에 따라 명령어가 달라집니다.

**Bash / CMD**
```bash
docker compose run --rm analyze analyze ./code/mig.sql -o markdown > report.md
```

**PowerShell**
```powershell
docker compose run --rm analyze analyze ./code/mig.sql -o markdown | Out-File -FilePath report.md -Encoding utf8
```

---

## 3. 주요 플래그
| 플래그 | 기본값 | 설명 |
| :--- | :--- | :--- |
| `--sandbox` | - | 오프라인 모드. 경로는 `./code/`로 시작해야 합니다. |
| `--output` | `console` | `console`, `markdown`, `csv`. |
