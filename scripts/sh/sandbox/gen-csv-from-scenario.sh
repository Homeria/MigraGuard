#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <scenario_name> [output_path] [--force]"
    exit 1
fi

S_PATH="experiments/scenarios/$1"
if [ ! -f "$S_PATH" ]; then S_PATH="$1"; fi

# Smart Output Detection
OUT=$2
FORCE_FLAG=""

if [[ "$OUT" == "--force" ]]; then
    FORCE_FLAG="--force"
    OUT=""
fi

if [ -z "$OUT" ]; then
    mkdir -p experiments/reports/metrics
    base=$(basename "$S_PATH" .yaml)
    OUT="experiments/reports/metrics/${base}_metrics.csv"
fi

echo "📊 Generating Metrics CSV from $1..."

go run ./cmd/migraguard simulate --scenario "$S_PATH" --csv "$OUT" --no-db "$FORCE_FLAG"

if [ $? -eq 0 ]; then
    echo "[OK] Metrics CSV successfully generated at: $OUT"
fi
