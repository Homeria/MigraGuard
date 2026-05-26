#!/bin/bash
# 🛡️ MigraGuard L2 Orchestrator: Single Scenario Pipeline (Standalone Executable)
# Usage: ./scripts/sh/run_scenario_pipeline.sh <scenario_yaml_path>

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$MODULE_DIR/../../" && pwd)"

# 1. Guard Clauses
if [ -z "$1" ]; then
    echo "❌ [ERROR] Scenario YAML path is required."
    echo "Usage: $0 <scenario_yaml_path>"
    exit 1
fi

S_PATH="$1"
# Resolve to absolute path if relative
if [[ "$S_PATH" != /* ]]; then
    S_PATH="$PROJECT_ROOT/$S_PATH"
fi

if [ ! -f "$S_PATH" ]; then
    echo "❌ [ERROR] Scenario file not found: $S_PATH"
    exit 1
fi

echo "=================================================="
echo "🛡️  MigraGuard L2 Scenario Pipeline: $(basename "$S_PATH")"
echo "=================================================="

# 0. Define Unique runs package directory
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
RUN_DIR="$PROJECT_ROOT/experiments/reports/batch_runs/${TIMESTAMP}"
mkdir -p "$RUN_DIR/data"

# 0. Clean old py scripts
echo "🧹 [0/5] Cleaning temporary Python files inside reports..."
find "$PROJECT_ROOT/experiments/reports" -name "*.py" -delete 2>/dev/null || true

# 1. Simulate seed
bash "$PROJECT_ROOT/scripts/sh/modules/simulate_scenario.sh" "$S_PATH"
if [ $? -ne 0 ]; then
    echo "❌ [ERROR] Simulation stage failed."
    exit 1
fi

# Extract seed DB name
EXP_NAME=$(grep "experiment_name:" "$S_PATH" | awk -F': ' '{print $2}' | tr -d " '\"")
SEED_DB="$PROJECT_ROOT/${EXP_NAME}.db"

if [ ! -f "$SEED_DB" ]; then
    # Fallback to local experiments/data if moved
    SEED_DB="$PROJECT_ROOT/experiments/data/${EXP_NAME}.db"
fi

if [ ! -f "$SEED_DB" ]; then
    echo "❌ [ERROR] Seed database not found: $SEED_DB"
    exit 1
fi

# 2. Dispatch DB to each case nested inside the runs pack (using override parameter)
echo "📂 [2/5] Dispatching databases to capacity cases nested in runs pack..."
configs=("$PROJECT_ROOT"/experiments/configs/cases/*.yaml)
for config in "${configs[@]}"; do
    if [ ! -f "$config" ]; then continue; fi
    case_name=$(grep -A 3 "metadata:" "$config" | grep "case_name:" | awk -F': ' '{print $2}' | tr -d " '\"")
    dest_db="$RUN_DIR/data/${case_name}.db"
    bash "$PROJECT_ROOT/scripts/sh/modules/dispatch_db.sh" "$SEED_DB" "$config" "$dest_db"
done

# Cleanup root seed db copy if it exists to keep workspace tidy
rm -f "$PROJECT_ROOT/${EXP_NAME}.db"

# 3. Batch Analyze (All Case configs x All DDL sqls) targeting runs DBs and local runs CSV inside data/
echo "🚀 [3/5] Executing dynamic batch analyses..."
REPORT_CSV="$RUN_DIR/data/research_results.csv"
rm -f "$REPORT_CSV"

sqls=("$PROJECT_ROOT"/experiments/ddl/*.sql)
count=0

for config in "${configs[@]}"; do
    if [ ! -f "$config" ]; then continue; fi
    case_name=$(grep -A 3 "metadata:" "$config" | grep "case_name:" | awk -F': ' '{print $2}' | tr -d " '\"")
    db_path="$RUN_DIR/data/${case_name}.db"
    
    for sql in "${sqls[@]}"; do
        ((count++))
        APPEND_FLAG=""
        if [ $count -gt 1 ]; then APPEND_FLAG="--append"; fi
        
        bash "$PROJECT_ROOT/scripts/sh/modules/analyze_single_ddl.sh" "$sql" "$db_path" "$config" "$REPORT_CSV" $APPEND_FLAG
    done
done

# 4. Generate heatmaps directly inside the runs pack
echo "📊 [4/5] Plotting 24h Individual heatmaps directly inside runs pack..."
python3 "$PROJECT_ROOT/tools/visualization/analyze/run_massive_forecast_plots.py" "$RUN_DIR"

# 5. Plot Master report directly inside the data/ folder of runs pack
echo "🎨 [5/5] Generating Master Heatmap Report directly inside data/ folder..."
python3 "$PROJECT_ROOT/tools/visualization/analyze/plot_research_report.py" "$REPORT_CSV"

# Sync results to artifact path for system integration
cp "$RUN_DIR/data/research_results_analysis.png" /home/gyeongho/.gemini/antigravity-cli/brain/fc61e4a8-7c53-4959-8257-7d147633cdc3/

echo "=================================================="
echo "✅ [L2 SUCCESS] Scenario Pipeline Complete!"
echo "📍 Master Chart: $RUN_DIR/data/research_results_analysis.png"
echo "=================================================="
exit 0


