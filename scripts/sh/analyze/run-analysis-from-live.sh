#!/bin/bash
if [ -z "$1" ]; then echo "Usage: $0 <sql_name> [db_url] [sqlite_path] [output_type]"; exit 1; fi
Q="experiments/ddl/$1"
if [ ! -f "$Q" ]; then Q="$1"; fi
DB_URL=$2
SQLITE=$3
OUT_TYPE=${4:-"console"}

ARGS="analyze $Q"
if [ ! -z "$DB_URL" ]; then ARGS="$ARGS --db $DB_URL"; fi
if [ ! -z "$SQLITE" ]; then ARGS="$ARGS --sqlite $SQLITE"; fi
if [ "$OUT_TYPE" != "console" ]; then ARGS="$ARGS --output $OUT_TYPE"; fi

go run ./cmd/migraguard $ARGS
