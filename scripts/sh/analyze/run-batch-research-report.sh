#!/bin/bash
REPORT="experiments/reports/research_results.csv"
mkdir -p experiments/reports
echo "🚀 Starting Massive Batch Research Analysis..."
count=0
dbs=(experiments/data/*.db)
sqls=(experiments/ddl/*.sql)

for db in "${dbs[@]}"; do
    for sql in "${sqls[@]}"; do
        ((count++))
        HEADER_FLAG=""
        if [ $count -gt 1 ]; then HEADER_FLAG="--no-header"; fi
        echo "[$count] $(basename $sql) @ $(basename $db)"
        if [ $count -eq 1 ]; then
            go run ./cmd/migraguard analyze "$sql" --sandbox "$db" --output csv $HEADER_FLAG > "$REPORT"
        else
            go run ./cmd/migraguard analyze "$sql" --sandbox "$db" --output csv $HEADER_FLAG >> "$REPORT"
        fi
    done
done
echo "✅ Research complete. Master report: $REPORT"
