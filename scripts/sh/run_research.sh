#!/bin/bash

# MigraGuard Batch Research Runner (Linux/Ubuntu)
# Executes all DDL cases against all scenarios and accumulates results to CSV

DATA_DIR="experiments/data"
DDL_DIR="experiments/ddl"
REPORT_FILE="experiments/reports/research_results.csv"

mkdir -p "experiments/reports"

echo "🚀 Starting Massive Batch Research Analysis..."
echo "📊 Target Report: $REPORT_FILE"

# 1. Initialize CSV with Header
FIRST_DDL=$(ls $DDL_DIR/*.sql | head -n 1)
FIRST_DB=$(ls $DATA_DIR/*.db | head -n 1)

if [ -z "$FIRST_DDL" ] || [ -z "$FIRST_DB" ]; then
    echo "[ERROR] No DDL or DB files found. Run seed_all first."
    exit 1
fi

go run ./cmd/migraguard analyze "$FIRST_DDL" --sandbox "$FIRST_DB" --output csv > "$REPORT_FILE"

# 2. Actual Loop
count=0
for db in $DATA_DIR/*.db; do
    for ddl in $DDL_DIR/*.sql; do
        ((count++))
        echo "[$count] Analyzing $(basename "$ddl") against $(basename "$db")..."
        
        go run ./cmd/migraguard analyze "$ddl" --sandbox "$db" --output csv --no-header >> "$REPORT_FILE"
    done
done

echo "--------------------------------------------------"
echo "✅ Research complete. $count cases processed."
echo "📈 Data saved to $REPORT_FILE"
