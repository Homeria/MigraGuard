#!/bin/bash

# MigraGuard One-Click Sandbox Script (Linux/Ubuntu)
# Usage: ./scripts/sandbox.sh <scenario_name> <ddl_name> [--docker] [--force]

SCENARIO_IN=$1
SQL_IN=$2
DOCKER=false
FORCE=false

# Parse flags
for arg in "$@"; do
  if [ "$arg" == "--docker" ]; then DOCKER=true; fi
  if [ "$arg" == "--force" ]; then FORCE=true; fi
done

if [ -z "$SCENARIO_IN" ] || [ -z "$SQL_IN" ]; then
  echo "Usage: $0 <scenario_name.yaml> <ddl_name.sql> [--docker] [--force]"
  exit 1
fi

# Resolve Paths
SCENARIO="experiments/scenarios/$SCENARIO_IN"
if [[ ! -f "$SCENARIO" ]]; then SCENARIO="$SCENARIO_IN"; fi # Fallback to literal path

SQL="experiments/ddl/$SQL_IN"
if [[ ! -f "$SQL" ]]; then SQL="$SQL_IN"; fi # Fallback to literal path

# 1. Extract Experiment Name from YAML
DB_NAME=$(grep "experiment_name" "$SCENARIO" | awk -F': ' '{print $2}' | tr -d '"' | tr -d "'" | tr -d '\r')
DB_PATH="experiments/data/${DB_NAME}.db"

echo "--------------------------------------------------"
echo "🚀 MigraGuard Sandbox Runner"
echo "Scenario: $SCENARIO"
echo "SQL:      $SQL"
echo "Sandbox:  $DB_PATH"
echo "--------------------------------------------------"

if [ "$DOCKER" = true ]; then
  echo "[MODE] Docker Container"
  FORCE_FLAG=""
  if [ "$FORCE" = true ]; then FORCE_FLAG="--force"; fi
  
  docker compose run --rm analyze-shell sh -c \
    "go run ./cmd/migraguard simulate --scenario $SCENARIO $FORCE_FLAG && \
     go run ./cmd/migraguard analyze $SQL --sandbox $DB_PATH"
else
  echo "[MODE] Local Execution"
  
  SIM_CMD="go run ./cmd/migraguard simulate --scenario $SCENARIO"
  if [ "$FORCE" = true ]; then SIM_CMD="$SIM_CMD --force"; fi
  
  $SIM_CMD && go run ./cmd/migraguard analyze "$SQL" --sandbox "$DB_PATH"
fi
