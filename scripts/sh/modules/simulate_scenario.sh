#!/bin/bash
# 🛡️ MigraGuard L1 Module: Simulate Scenario (Standalone Executable)
# Usage: ./scripts/sh/modules/simulate_scenario.sh <scenario_yaml_path>

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$MODULE_DIR/../../../" && pwd)"

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

# 2. Run simulation
echo "🔄 [L1] Simulating Scenario: $(basename "$S_PATH")"
cd "$PROJECT_ROOT" || exit 1

./build/migraguard simulate --scenario "$S_PATH" --force
if [ $? -ne 0 ]; then
    echo "❌ [ERROR] Simulation execution failed."
    exit 1
fi

echo "✅ [L1 SUCCESS] Simulation complete for $(basename "$S_PATH")"
exit 0
