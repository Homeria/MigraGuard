# 🖥️ Native: 00. Setup & Build

Build MigraGuard binaries and prepare your environment.

---

## 1. Build Binaries
Requires Go 1.25+.

**Unix / PowerShell**
```bash
go build -o ./migraguard ./cmd/migraguard
go build -o ./loadgen ./cmd/loadgen
```

**Windows CMD**
```cmd
go build -o migraguard.exe ./cmd/migraguard
go build -o loadgen.exe ./cmd/loadgen
```

## 2. Configuration
Copy `migraguard.yaml.example` to `migraguard.yaml` and update your PostgreSQL DSN.
