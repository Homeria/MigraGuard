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

---

## 4. 고도화된 일괄 오케스트레이션 파이프라인 (Orchestration Scripts)

v3.9에서 구축된 Standalone L1 마이크로 모듈들과 L2/L3 오케스트레이터를 통해, 복잡한 다차원 장비 스펙 및 DDL의 조합 시뮬레이션을 명령어 한 줄로 전수 자동화 구동할 수 있습니다. 각 플랫폼(Linux Bash, Windows PowerShell, Windows CMD Batch)별로 완벽히 대칭 설계되었습니다.

### A. Linux Bash 파이프라인

- **L1 Standalone 모듈**:
  - `bash scripts/sh/modules/simulate_scenario.sh <scenario_yaml>`
  - `bash scripts/sh/modules/dispatch_db.sh <seed_db> <case_yaml>`
  - `bash scripts/sh/modules/analyze_single_ddl.sh <ddl> <db> <config> <out_csv>`
- **L2 단일 시나리오 파이프라인**:
  ```bash
  bash scripts/sh/run_scenario_pipeline.sh experiments/scenarios/02_commuter_daily_rush.yaml
  ```
- **L3 글로벌 전수 오케스트레이터**:
  ```bash
  bash scripts/sh/run_global_pipeline.sh
  ```

### B. Windows PowerShell 파이프라인

- **L1 Standalone 모듈**:
  - `.\scripts\ps1\modules\simulate_scenario.ps1 <scenario_yaml>`
  - `.\scripts\ps1\modules\dispatch_db.ps1 <seed_db> <case_yaml>`
  - `.\scripts\ps1\modules\analyze_single_ddl.ps1 <ddl> <db> <config> <out_csv>`
- **L2 단일 시나리오 파이프라인**:
  ```powershell
  .\scripts\ps1\run_scenario_pipeline.ps1 .\experiments\scenarios\02_commuter_daily_rush.yaml
  ```
- **L3 글로벌 전수 오케스트레이터**:
  ```powershell
  .\scripts\ps1\run_global_pipeline.ps1
  ```

### C. Windows CMD Batch 파이프라인

- **L1 Standalone 모듈**:
  - `call .\scripts\cmd\modules\simulate_scenario.cmd <scenario_yaml>`
  - `call .\scripts\cmd\modules\dispatch_db.cmd <seed_db> <case_yaml>`
  - `call .\scripts\cmd\modules\analyze_single_ddl.cmd <ddl> <db> <config> <out_csv>`
- **L2 단일 시나리오 파이프라인**:
  ```cmd
  call .\scripts\cmd\run_scenario_pipeline.cmd experiments\scenarios\02_commuter_daily_rush.yaml
  ```
- **L3 글로벌 전수 오케스트레이터**:
  ```cmd
  call .\scripts\cmd\run_global_pipeline.cmd
  ```

