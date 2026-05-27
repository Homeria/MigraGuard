#!/bin/bash
# 🛡️ MigraGuard L1 Module: Dispatch DB to Case (Standalone Executable)
# Usage: ./scripts/sh/modules/dispatch_db.sh <seed_db_path> <case_yaml_path> [override_db_path]

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$MODULE_DIR/../../../" && pwd)"

# 1. Guard Clauses
if [ -z "$1" ] || [ -z "$2" ]; then
    echo "❌ [ERROR] Both seed_db_path and case_yaml_path are required."
    echo "Usage: $0 <seed_db_path> <case_yaml_path> [override_db_path]"
    exit 1
fi

SEED_DB="$1"
CASE_YAML="$2"
OVERRIDE_DB="$3"

# Resolve to absolute paths if relative
if [[ "$SEED_DB" != /* ]]; then SEED_DB="$PROJECT_ROOT/$SEED_DB"; fi
if [[ "$CASE_YAML" != /* ]]; then CASE_YAML="$PROJECT_ROOT/$CASE_YAML"; fi

if [ ! -f "$SEED_DB" ]; then
    echo "❌ [ERROR] Seed DB not found: $SEED_DB"
    exit 1
fi
if [ ! -f "$CASE_YAML" ]; then
    echo "❌ [ERROR] Case YAML configuration not found: $CASE_YAML"
    exit 1
fi

# 2. Extract db_path and case_name dynamically from case configuration metadata block
CASE_NAME=$(grep -A 3 "metadata:" "$CASE_YAML" | grep "case_name:" | awk -F': ' '{print $2}' | tr -d " '\"")

DB_PATH=""
if [ -n "$OVERRIDE_DB" ]; then
    DB_PATH="$OVERRIDE_DB"
else
    DB_PATH=$(grep -A 3 "metadata:" "$CASE_YAML" | grep "db_path:" | awk -F': ' '{print $2}' | tr -d " '\"")
fi

if [ -z "$DB_PATH" ]; then
    echo "❌ [ERROR] Could not parse db_path from configuration metadata: $CASE_YAML"
    exit 1
fi

# Resolve destination db_path to absolute path if relative
if [[ "$DB_PATH" != /* ]]; then
    DB_PATH="$PROJECT_ROOT/$DB_PATH"
fi

# 3. Create destination directory and copy database
mkdir -p "$(dirname "$DB_PATH")"
cp "$SEED_DB" "$DB_PATH"
if [ $? -ne 0 ]; then
    echo "❌ [ERROR] Failed to dispatch database to $DB_PATH"
    exit 1
fi

echo "   -> [L1 SUCCESS] Dispatched to Case [$CASE_NAME]: $(basename "$DB_PATH")"
exit 0

