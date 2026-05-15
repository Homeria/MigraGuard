# 🖥️ 네이티브: 04. 샌드박스 내보내기 (Export Sandbox)

Export Sandbox 명령은 오프라인 시뮬레이션 파일에 저장된 모든 메트릭을 구조화된 CSV로 추출합니다.

---

## 1. 실행 예시

**Bash**
```bash
./migraguard export-sandbox -i scenario.db -o output.csv
```

**PowerShell**
```powershell
.\migraguard.exe export-sandbox -i scenario.db -o output.csv
```

**CMD**
```cmd
migraguard.exe export-sandbox -i scenario.db -o output.csv
```

---

## 2. 상세 플래그 참조

| 플래그 | 약어 | 기본값 | 설명 |
| :--- | :--- | :--- | :--- |
| `--input` | `-i` | **(필수)** | **소스 데이터베이스**. `simulate` 명령으로 생성된 SQLite 샌드박스 파일 경로입니다. |
| `--output` | `-o` | **(필수)** | **대상 CSV 경로**. 추출된 메트릭이 저장될 파일 경로입니다. |

---

## 3. CSV 데이터 필드
내보낸 CSV에는 `Timestamp`, `TableName`, `TableSize`, `ReplicationLag`, `ActiveConnections`, `P99Time`, `TPS`, `SharedBlksHit`, `SharedBlksRead` 필드가 포함됩니다.
