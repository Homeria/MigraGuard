#!/bin/bash
# 🛡️ MigraGuard: Backward Compatibility Wrapper for run_full_workflow.sh
# Relocated to L2 Orchestrator: scripts/sh/run_scenario_pipeline.sh

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$MODULE_DIR/../../../" && pwd)"

# Redirect calls to the new L2 modular orchestrator script
bash "$PROJECT_ROOT/scripts/sh/run_scenario_pipeline.sh" "$@"
exit $?

