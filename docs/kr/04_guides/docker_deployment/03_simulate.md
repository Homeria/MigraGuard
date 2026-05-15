# 🐳 Docker: 03. 시뮬레이션 (Simulate)

결정론적인 워크로드 시나리오를 생성합니다.

---

## 1. 실행 명령어 (통합)
명령어는 모든 쉘에서 동일합니다.
```bash
docker compose run --rm analyze simulate --scenario ./code/experiments/scenarios/03_spike_flash_sale.yaml
```

## 2. 호스트로 직접 CSV 추출
리다이렉션을 사용하여 로우 데이터를 저장할 때의 차이점입니다.

**Bash / CMD**
```bash
docker compose run --rm analyze simulate -s ./code/scen.yaml --csv - > host_metrics.csv
```

**PowerShell**
```powershell
docker compose run --rm analyze simulate -s ./code/scen.yaml --csv - | Out-File -FilePath host_metrics.csv -Encoding utf8
```

---

## 3. 주요 플래그
| 플래그 | 기본값 | 설명 |
| :--- | :--- | :--- |
| `--scenario` | (필수) | 입력 YAML 경로 (`./code/`로 시작). |
| `--csv` | - | 내보내기 경로. `-` 입력 시 표준 출력 리다이렉션 사용. |
