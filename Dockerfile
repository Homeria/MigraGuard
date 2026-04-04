# --- Build Stage ---
FROM golang:1.25-bookworm AS builder

# Install build essentials for CGO (required by pg_query_go)
RUN apt-get update && apt-get install -y build-essential libssl-dev

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the source code
COPY . .

# Build the migraguard binary
# Using CGO_ENABLED=1 because pg_query_go needs CGO
RUN CGO_ENABLED=1 GOOS=linux go build -o /migraguard ./cmd/migraguard

# --- Final Stage ---
FROM debian:bookworm-slim

# Install runtime dependencies for CGO and SQLite
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /migraguard /app/migraguard

# Create data directory for SQLite
RUN mkdir -p /app/data && chmod 777 /app/data

# Default SQLite path (can be overridden by config or flags)
ENV SQLITE_PATH="/app/data/migraguard.db"

# Expose no ports as this is a CLI/Agent tool
# ENTRYPOINT will be overridden in docker-compose for different modes
ENTRYPOINT ["/app/migraguard"]
