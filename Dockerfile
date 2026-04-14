# --- Build Stage ---
FROM golang:1.25-bookworm AS builder

# CGO 및 pg_query_go를 위한 빌드 도구 설치
RUN apt-get update && apt-get install -y build-essential libssl-dev

WORKDIR /app

# 의존성 복사 및 캐싱
COPY go.mod go.sum ./
RUN go mod tidy

# 전체 소스 코드 복사
COPY . .

# 1. 메인 도구 (MigraGuard) 빌드
RUN CGO_ENABLED=1 GOOS=linux go build -o /migraguard ./cmd/migraguard

# 2. 부하 생성기 (Load Generator) 빌드
RUN CGO_ENABLED=1 GOOS=linux go build -o /loadgen ./cmd/loadgen

# --- Final Stage ---
FROM debian:bookworm-slim

# 실행 환경 의존성 설치
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 빌더로부터 두 개의 바이너리 모두 복사
COPY --from=builder /migraguard /app/migraguard
COPY --from=builder /loadgen /app/loadgen

# SQLite 저장용 디렉토리 생성
RUN mkdir -p /app/data && chmod 777 /app/data

# 기본 환경 변수
ENV SQLITE_PATH="/app/data/migraguard.db"

# 기본 실행 파일 설정
ENTRYPOINT ["/app/migraguard"]
