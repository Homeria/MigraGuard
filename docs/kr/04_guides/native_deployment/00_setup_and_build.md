# 🖥️ 네이티브: 00. 설치 및 빌드 (Setup & Build)

MigraGuard 바이너리를 빌드하고 실행 환경을 준비합니다.

---

## 1. 바이너리 빌드
Go 1.25 버전 이상이 필요합니다.

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

## 2. 설정 (Configuration)
`migraguard.yaml.example`을 `migraguard.yaml`로 복사하고 PostgreSQL DSN 정보를 업데이트합니다.
