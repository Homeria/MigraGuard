#!/bin/bash
# 🛡️ MigraGuard L1 Module: Analyze Single DDL (Standalone Executable)
# Usage: ./scripts/sh/modules/analyze_single_ddl.sh <ddl_path> <db_path> <config_path> <output_csv_path> [--append]

MODULE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$MODULE_DIR/../../../" && pwd)"

# 1. Guard Clauses
if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ] || [ -z "$4" ]; then
    echo "❌ [ERROR] Missing required arguments."
    echo "Usage: $0 <ddl_path> <db_path> <config_path> <output_csv_path> [--append]"
    exit 1
fi

DDL_PATH="$1"
DB_PATH="$2"
CONFIG_PATH="$3"
OUT_CSV="$4"
MODE="OVERWRITE"

if [ "$5" = "--append" ]; then
    MODE="APPEND"
fi

# Resolve paths to absolute if relative
if [[ "$DDL_PATH" != /* ]]; then DDL_PATH="$PROJECT_ROOT/$DDL_PATH"; fi
if [[ "$DB_PATH" != /* ]]; then DB_PATH="$PROJECT_ROOT/$DB_PATH"; fi
if [[ "$CONFIG_PATH" != /* ]]; then CONFIG_PATH="$PROJECT_ROOT/$CONFIG_PATH"; fi
if [[ "$OUT_CSV" != /* ]]; then OUT_CSV="$PROJECT_ROOT/$OUT_CSV"; fi

if [ ! -f "$DDL_PATH" ]; then
    echo "❌ [ERROR] DDL file not found: $DDL_PATH"
    exit 1
fi
if [ ! -f "$DB_PATH" ]; then
    echo "❌ [ERROR] Database file not found: $DB_PATH"
    exit 1
fi
if [ ! -f "$CONFIG_PATH" ]; then
    echo "❌ [ERROR] Config file not found: $CONFIG_PATH"
    exit 1
fi

# 2. Run analysis
cd "$PROJECT_ROOT" || exit 1

HEADER_FLAG=""
if [ "$MODE" = "APPEND" ] && [ -f "$OUT_CSV" ]; then
    HEADER_FLAG="--no-header"
fi

mkdir -p "$(dirname "$OUT_CSV")"

if [ "$MODE" = "OVERWRITE" ]; then
    ./build/migraguard analyze "$DDL_PATH" --sandbox "$DB_PATH" --output csv $HEADER_FLAG --config "$CONFIG_PATH" > "$OUT_CSV" || true
else
    ./build/migraguard analyze "$DDL_PATH" --sandbox "$DB_PATH" --output csv $HEADER_FLAG --config "$CONFIG_PATH" >> "$OUT_CSV" || true
fi

echo "   -> [L1 SUCCESS] Analyzed $(basename "$DDL_PATH") on $(basename "$DB_PATH") (Mode: $MODE)"
exit 0
