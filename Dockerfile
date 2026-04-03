FROM golang:1.25-bookworm

# Install build essentials for CGO (required by pg_query_go)
RUN apt-get update && apt-get install -y build-essential libssl-dev

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the source code
COPY . .

# Build the migraguard binary
RUN go build -o /migraguard ./cmd/migraguard

# Ensure the binary is executable
RUN chmod +x /migraguard

ENTRYPOINT ["/migraguard"]
