#!/bin/bash

# MigraGuard Batch Seeder (Linux/Ubuntu)
# Seeds all scenarios found in experiments/scenarios/ into experiments/data/

DATA_DIR="experiments/data"
SCENARIO_DIR="experiments/scenarios"

mkdir -p "$DATA_DIR"

echo "🚀 Starting Batch Seeding for all scenarios..."

for scenario in "$SCENARIO_DIR"/*.yaml; do
    if [ -f "$scenario" ]; then
        filename=$(basename "$scenario")
        echo "--------------------------------------------------"
        echo "Processing: $filename"
        
        # We run simulate. The engine will place the .db in experiments/data 
        # because our scripts/sandbox.sh and app logic use the experiment_name.
        # To ensure it goes to experiments/data, we run from root and 
        # the app will create it in the current dir, then we move it if needed,
        # OR we rely on the fact that simulate command now places it relative to execution.
        
        go run ./cmd/migraguard simulate --scenario "$scenario" --force
        
        # After simulation, the DB is created in the root (default behavior of app)
        # We move it to experiments/data/ for organization
        DB_NAME=$(grep "experiment_name" "$scenario" | awk -F': ' '{print $2}' | tr -d '"' | tr -d "'" | tr -d '\r')
        if [ -f "${DB_NAME}.db" ]; then
            mv "${DB_NAME}.db" "$DATA_DIR/"
            echo "[OK] Generated: $DATA_DIR/${DB_NAME}.db"
        fi
    fi
done

echo "--------------------------------------------------"
echo "✅ Batch seeding complete. All databases are in $DATA_DIR"
