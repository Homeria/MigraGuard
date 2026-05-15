# 🖥️ 네이티브: 03. 시뮬레이션 (Simulate)

Simulate 명령은 연구 및 테스트를 위해 수학적 모델을 기반으로 가상의 시계열 메트릭을 생성합니다.

---

## 1. 실행 예시

**Bash**
```bash
./migraguard simulate --scenario experiments/scenarios/03_spike_flash_sale.yaml
```

**PowerShell**
```powershell
.\migraguard.exe simulate --scenario .\experiments\scenarios\03_spike_flash_sale.yaml
```

**CMD**
```cmd
migraguard.exe simulate --scenario experiments\scenarios\03_spike_flash_sale.yaml
```

---

## 2. 상세 플래그 참조

| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--scenario` | `-s` | **(필수)** | **입력 시나리오 경로**. 워크로드 프로필(TPS, 이벤트, 스파이크 등)을 정의하는 YAML 파일을 가리킵니다. |
| `--force` | `-f` | `false` | **강제 덮어쓰기**. 대상 샌드박스 데이터베이스가 이미 존재하는 경우, 이 플래그를 사용하여 덮어쓸 수 있습니다. |
| `--csv` | - | - | **CSV 내보내기 경로**. 지정된 경우 생성된 메트릭을 이 CSV 파일로도 내보냅니다. 표준 출력은 `-`를 사용하십시오. |
| `--no-db` | - | `false` | **일회성 모드**. `--csv`와 함께 사용하면 CSV 내보내기 후 생성된 SQLite 파일을 삭제합니다. |

---

## 3. CSV 데이터 저장 (리다이렉션)

**Bash / CMD**
```bash
./migraguard simulate -s scenario.yaml --csv - > metrics.csv
```

**PowerShell**
```powershell
.\migraguard.exe simulate -s scenario.yaml --csv - | Out-File -FilePath metrics.csv -Encoding utf8
```
