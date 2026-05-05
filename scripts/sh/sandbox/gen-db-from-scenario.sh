#!/bin/bash
if [ -z "$1" ]; then echo "Usage: $0 <scenario_name> [--force]"; exit 1; fi
S="experiments/scenarios/$1"
if [ ! -f "$S" ]; then S="$1"; fi
go run ./cmd/migraguard simulate --scenario "$S" "$2"
N=$(grep "experiment_name:" "$S" | awk -F': ' '{print $2}' | tr -d " '\"")
if [ -f "${N}.db" ]; then
    mkdir -p experiments/data/
    mv "${N}.db" experiments/data/
    echo "[OK] Generated: experiments/data/${N}.db"
fi
