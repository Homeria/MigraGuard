# 🐳 Docker: 03. 시뮬레이션 (Simulate)

결정론적인 워크로드 시나리오를 생성합니다.

---

## 1. 실행 명령어 (통합)
명령어는 모든 쉘에서 동일합니다.
```bash
docker compose run --rm analyze simulate --scenario ./code/experiments/scenarios/05_spike_flash_sale.yaml
```

생성되는 DB 파일명은 시나리오 YAML의 `experiment_name` 값을 따릅니다.

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
| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--scenario` | `-s` | **(필수)** | 입력 YAML 경로 (`./code/`로 시작). |
| `--force` | `-f` | `false` | 대상 데이터베이스가 이미 존재하는 경우, 덮어쓰기를 강제합니다. |
| `--csv` | - | - | 내보내기 경로. `-` 입력 시 표준 출력 리다이렉션 사용. |
| `--no-db` | - | `false` | 일회성 모드. `--csv`와 함께 사용 시 CSV 출력 후 SQLite 파일을 자동 제거합니다. |

주의: `--no-db`는 `--csv`와 함께 사용할 때만 DB 삭제 동작을 수행합니다.
