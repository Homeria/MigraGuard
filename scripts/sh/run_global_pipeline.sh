#!/bin/bash
# 🛡️ MigraGuard L3 Orchestrator: Global Multi-Scenario Pipeline (Standalone Executable)
# Usage: ./scripts/sh/run_global_pipeline.sh

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$MODULE_DIR/../../" && pwd)"

echo "=================================================="
echo "🛡️  MigraGuard L3 Global Multi-Scenario Orchestrator"
echo "=================================================="

# 1. Collect all scenarios
scenarios=("$PROJECT_ROOT"/experiments/scenarios/*.yaml)
total=${#scenarios[@]}

if [ "$total" -eq 0 ] || [ ! -f "${scenarios[0]}" ]; then
    echo "❌ [ERROR] No scenario files found in experiments/scenarios/"
    exit 1
fi

echo "📂 Found $total scenario file(s) to process."
echo "=================================================="

count=0
for scenario in "${scenarios[@]}"; do
    if [ ! -f "$scenario" ]; then continue; fi
    ((count++))
    s_name=$(basename "$scenario")
    
    echo ""
    echo "=================================================="
    echo "👉 [Scenario $count/$total] Processing: $s_name"
    echo "=================================================="
    
    # Call L2 Orchestrator for this specific scenario
    bash "$PROJECT_ROOT/scripts/sh/run_scenario_pipeline.sh" "$scenario"
    
    if [ $? -ne 0 ]; then
        echo "⚠️  [WARNING] Pipeline execution failed for scenario: $s_name. Continuing to next."
    else
        echo "✅ [Scenario SUCCESS] Completed scenario: $s_name"
    fi
done

echo ""
echo "=================================================="
echo "🏆 [L3 SUCCESS] Global Multi-Scenario Pipeline Completed!"
echo "📂 All archived outputs are located in experiments/reports/batch_runs/"
echo "=================================================="
exit 0
