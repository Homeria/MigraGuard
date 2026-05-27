#!/bin/bash
if [ -z "$1" ] || [ -z "$2" ]; then echo "Usage: $0 <db_name> <sql_name> [output_type]"; exit 1; fi
D="experiments/data/$1"
if [ ! -f "$D" ]; then D="$1"; fi
Q="experiments/ddl/$2"
if [ ! -f "$Q" ]; then Q="$2"; fi
O=${3:-"console"}
go run ./cmd/migraguard analyze "$Q" --sandbox "$D" --output "$O"
