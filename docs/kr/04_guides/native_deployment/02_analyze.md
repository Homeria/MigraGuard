# 🖥️ 네이티브: 02. 리스크 분석 (Analyze)

Analyze 명령은 SQL 시맨틱과 워크로드 메트릭을 교차 참조하여 DDL 마이그레이션의 리스크를 평가하는 핵심 엔진입니다.

---

## 1. 실행 예시 (Live 모드)

**Bash**
```bash
./migraguard analyze migration.sql
```

**PowerShell**
```powershell
.\migraguard.exe analyze migration.sql
```

**CMD**
```cmd
migraguard.exe analyze migration.sql
```

---

## 2. 실행 예시 (Sandbox / 오프라인 모드)

**Bash**
```bash
./migraguard analyze migration.sql --sandbox experiments/data/01_steady_normal.db
```

**PowerShell**
```powershell
.\migraguard.exe analyze migration.sql --sandbox experiments\data\01_steady_normal.db
```

**CMD**
```cmd
migraguard.exe analyze migration.sql --sandbox experiments\data\01_steady_normal.db
```

---

## 3. 리포트 저장 (Markdown 형식)

**Bash / CMD**
```bash
./migraguard analyze migration.sql -o markdown > report.md
```

**PowerShell**
```powershell
.\migraguard.exe analyze migration.sql -o markdown | Out-File -FilePath report.md -Encoding utf8
```

---

## 4. 상세 플래그 참조

| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--db` | - | (설정 참조) | **PostgreSQL DSN**. Live 모드에서 현재 테이블 통계(크기, 인덱스 상태 등)를 가져오는 데 사용됩니다. |
| `--sqlite` | - | `migraguard.db` | **메트릭 저장소 경로**. 에이전트의 데이터베이스를 가리키며, 과거 TPS 기준점을 검색하는 데 사용됩니다. |
| `--sandbox` | `-s` | - | **시뮬레이션 데이터베이스 경로**. 이 플래그를 설정하면 **오프라인 모드**가 활성화되어 모든 실시간 DB 연결을 무시합니다. |
| `--output` | `-o` | `console` | **출력 포맷**. `console`(색상 텍스트), `markdown`(포맷된 보고서), `csv`(로우 데이터) 중에서 선택할 수 있습니다. |
| `--no-header` | - | `false` | **CSV 헤더 억제**. `-o csv` 사용 시 헤더 행이 출력되지 않도록 합니다. |
| `--forecast` | `-f` | `false` | **예측 엔진 활성화**. 24시간 베이스라인을 기준으로 리스크를 분석하고 골든 윈도우를 식별합니다. |

---

## 5. 종료 코드 (Exit Codes)
*   **`0`**: 마이그레이션이 **Safe** 또는 **Warning** 수준으로 간주됩니다.
*   **`1`**: 마이그레이션이 **Danger** 등급(리스크 점수 > 80)이거나 내부 오류가 발생했습니다.
� DB 연결을 무시합니다. |
| `--output` | `-o` | `console` | **출력 포맷**. `console`(색상 텍스트), `markdown`(포맷된 보고서), `csv`(로우 데이터) 중에서 선택할 수 있습니다. |
| `--no-header` | - | `false` | **CSV 헤더 억제**. `-o csv` 사용 시 헤더 행이 출력되지 않도록 합니다. |

---

## 5. 종료 코드 (Exit Codes)
*   **`0`**: 마이그레이션이 **Safe** 또는 **Warning** 수준으로 간주됩니다.
*   **`1`**: 마이그레이션이 **Danger** 등급(리스크 점수 > 80)이거나 내부 오류가 발생했습니다.
