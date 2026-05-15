# 🖥️ Native: 02. Analyze (Risk Assessment)

The Analyze command is the core engine that evaluates the risk of a DDL migration by cross-referencing SQL semantics with workload metrics.

---

## 1. Execution Examples (Live Mode)

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

## 2. Execution Examples (Sandbox / Offline Mode)

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

## 3. Saving Reports (Markdown)

**Bash / CMD**
```bash
./migraguard analyze migration.sql -o markdown > report.md
```

**PowerShell**
```powershell
.\migraguard.exe analyze migration.sql -o markdown | Out-File -FilePath report.md -Encoding utf8
```

---

## 4. Detailed Flag Reference

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--db` | - | (From Config) | **PostgreSQL DSN**. Used in Live mode to fetch current table statistics (size, index status). |
| `--sqlite` | - | `migraguard.db` | **Metrics Storage Path**. Points to the agent's database to retrieve historical TPS baselines. |
| `--sandbox` | `-s` | - | **Simulation Database Path**. Setting this flag triggers **Offline Mode**, ignoring any live DB connection. |
| `--output` | `-o` | `console` | **Output Format**. Choose between `console` (colored text), `markdown` (formatted reports), or `csv` (raw data). |
| `--no-header` | - | `false` | **Suppress CSV Header**. If using `-o csv`, this flag prevents the header row from being printed. |

---

## 5. Exit Codes
*   **`0`**: Migration is considered **Safe** or **Warning**.
*   **`1`**: Migration is **Danger** (Risk score > 80) or an internal error occurred.
