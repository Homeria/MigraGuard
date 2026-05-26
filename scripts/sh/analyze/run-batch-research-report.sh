#!/bin/bash
REPORT="experiments/reports/research_results.csv"
mkdir -p experiments/reports
echo "🚀 Starting Massive Batch Research Analysis..."
count=0
sqls=(experiments/ddl/*.sql)

# Clean previous report to ensure freshness
rm -f "$REPORT"

configs=(experiments/configs/cases/*.yaml)

for config in "${configs[@]}"; do
    if [ ! -f "$config" ]; then continue; fi
    
    # Dynamically extract db_path and case_name from the YAML metadata block
    db_path=$(grep -A 3 "metadata:" "$config" | grep "db_path:" | awk -F': ' '{print $2}' | tr -d " '\"")
    case_name=$(grep -A 3 "metadata:" "$config" | grep "case_name:" | awk -F': ' '{print $2}' | tr -d " '\"")
    
    if [ -z "$db_path" ]; then
        echo "⚠️  [WARNING] Could not parse db_path from $config. Skipping."
        continue
    fi
    
    if [ ! -f "$db_path" ]; then
        echo "⚠️  [WARNING] DB file not found for case [$case_name]: $db_path. Skipping."
        continue
    fi

    echo "=================================================="
    echo "🔥 Processing Case: [$case_name] using config: $config"
    echo "=================================================="

    for sql in "${sqls[@]}"; do
        ((count++))
        HEADER_FLAG=""
        if [ $count -gt 1 ]; then HEADER_FLAG="--no-header"; fi
        echo "[$count] $(basename $sql) @ $(basename $db_path)"
        
        # We must use the pre-built build/migraguard binary for super fast batch run.
        # Danger exits with 1, we must ignore the exit code (|| true) to let the loop continue.
        if [ $count -eq 1 ]; then
            ./build/migraguard analyze "$sql" --sandbox "$db_path" --output csv $HEADER_FLAG --config "$config" > "$REPORT" || true
        else
            ./build/migraguard analyze "$sql" --sandbox "$db_path" --output csv $HEADER_FLAG --config "$config" >> "$REPORT" || true
        fi
    done
done

echo "✅ Research complete. Master report: $REPORT"
