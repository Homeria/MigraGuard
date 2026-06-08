# 네이티브: 02. 리스크 분석 (Analyze)

`analyze` 명령은 마이그레이션 SQL 파일을 파싱하고, 라이브 PostgreSQL 지표 또는 SQLite 샌드박스 지표와 결합해 DDL 위험도를 평가합니다.

---

## 1. Live 모드

Live 모드는 실제 PostgreSQL 연결에서 테이블 크기, active connection, `pg_stat_statements` 기반 지표를 조회하고, 로컬 SQLite에 저장된 과거 workload를 함께 사용합니다.

**Bash**
```bash
./migraguard analyze migration.sql --db "postgres://user:pass@host:5432/db" --sqlite migraguard.db
```

**PowerShell**
```powershell
.\migraguard.exe analyze migration.sql --db "postgres://user:pass@host:5432/db" --sqlite .\migraguard.db
```

`--db`와 `--sqlite`를 생략하면 `--config`로 지정한 설정 파일 또는 기본값을 사용합니다.

---

## 2. Sandbox / Offline 모드

Sandbox 모드는 실제 PostgreSQL에 연결하지 않고, `simulate`로 생성한 SQLite 샌드박스를 가상 PostgreSQL 어댑터처럼 사용합니다.

**Bash**
```bash
./migraguard analyze migration.sql --sandbox exp_05_spike_flash_sale.db
```

**PowerShell**
```powershell
.\migraguard.exe analyze migration.sql --sandbox .\exp_05_spike_flash_sale.db
```

`--sandbox`를 지정하면 `--db`는 무시됩니다.

---

## 3. 리포트 출력

**Markdown**
```bash
./migraguard analyze migration.sql --sandbox exp_05_spike_flash_sale.db --output markdown > report.md
```

**CSV**
```bash
./migraguard analyze migration.sql --sandbox exp_05_spike_flash_sale.db --output csv > result.csv
```

CSV를 여러 번 이어 붙일 때는 두 번째 실행부터 `--no-header`를 사용합니다.

```bash
./migraguard analyze migration.sql --sandbox exp_05_spike_flash_sale.db --output csv --no-header >> result.csv
```

---

## 4. 24시간 예측 분석

`--forecast`를 사용하면 SQLite에 저장된 시간대별 workload 패턴을 기반으로 24시간 위험도를 계산합니다.

```bash
./migraguard analyze migration.sql --sandbox exp_05_spike_flash_sale.db --forecast
```

주의:
- 현재 `--forecast`에는 짧은 옵션이 없습니다. `-f`는 지원하지 않습니다.
- 예측 결과가 생성되면 현재 작업 디렉터리에 `predictive_forecast.csv`가 생성됩니다.
- 현재 CLI에는 forecast CSV 경로를 바꾸는 옵션이 없습니다.

---

## 5. 상세 플래그 참조

| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--db` | - | 설정 참조 | Live 모드에서 사용할 PostgreSQL DSN입니다. |
| `--sqlite` | - | `migraguard.db` 또는 설정값 | Agent가 저장한 로컬 SQLite workload 저장소입니다. |
| `--sandbox` | `-s` | - | Offline 모드에서 사용할 SQLite 샌드박스 DB 경로입니다. 지정 시 실제 PostgreSQL 연결을 사용하지 않습니다. |
| `--output` | `-o` | `console` | 출력 형식입니다. 현재 `console`, `markdown`, `csv`를 의도합니다. |
| `--no-header` | - | `false` | `--output csv`에서 CSV 헤더를 출력하지 않습니다. |
| `--forecast` | - | `false` | 24시간 예측 분석을 활성화합니다. |

주의: 현재 구현은 `--output` 값이 잘못되어도 에러를 내지 않고 `console` 출력으로 처리합니다. 문서와 실험에서는 `console`, `markdown`, `csv` 중 하나만 사용하십시오.

---

## 6. 종료 코드

| 종료 코드 | 의미 |
| :--- | :--- |
| `0` | 분석 결과가 `Safe` 또는 `Warning`입니다. |
| `1` | 분석 결과가 `Danger`이거나, 초기화/파싱/분석 중 오류가 발생했습니다. |

CI/CD에서 사용할 때는 `Danger`로 인한 차단과 내부 오류가 모두 `1`로 표현된다는 점을 고려해야 합니다.
