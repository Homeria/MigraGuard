#!/bin/bash
mkdir -p experiments/reports/metrics
for f in experiments/scenarios/*.yaml; do
    basename=$(basename "$f" .yaml)
    echo "🔄 Exporting CSV for: $basename"
    "$(dirname "$0")/gen-csv-from-scenario.sh" "$f" "experiments/reports/metrics/${basename}_metrics.csv" "$1"
done
echo "✅ All metrics exported to experiments/reports/metrics/"
