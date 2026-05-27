#!/bin/bash
# 🛡️ MigraGuard: Predictive 24-Hour Forecast Workflow
# Usage: ./scripts/sh/analyze/run-predictive-forecast.sh <db_path> <ddl_path>

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: $0 <db_path> <ddl_path>"
    echo "Example: $0 experiments/data/exp_01_steady_normal.db experiments/ddl/011_danger_rewrite_order_no.sql"
    exit 1
fi

DB_PATH=$1
DDL_PATH=$2
CSV_NAME="predictive_forecast.csv"
IMG_NAME="predictive_risk_heatmap.png"

echo "--------------------------------------------------------"
echo "🚀 [1/2] Running Predictive Risk Analysis (Go Engine)..."
echo "--------------------------------------------------------"
go run ./cmd/migraguard analyze "$DDL_PATH" --sandbox "$DB_PATH" --forecast --output console

if [ $? -ne 0 ]; then
    echo "❌ [ERROR] Analysis failed. Skipping visualization."
    exit 1
fi

echo ""
echo "--------------------------------------------------------"
echo "📊 [2/2] Generating Risk Heatmap Visualization (Python)..."
echo "--------------------------------------------------------"
if [ ! -f "$CSV_NAME" ]; then
    echo "❌ [ERROR] Forecast CSV ($CSV_NAME) was not generated."
    exit 1
fi

python ./tools/visualization/analyze/plot_predictive_heatmap.py "$CSV_NAME"

if [ $? -eq 0 ]; then
    echo "✅ [SUCCESS] Analysis complete!"
    echo "📍 Report Location: $IMG_NAME"
else
    echo "❌ [ERROR] Visualization failed."
    exit 1
fi
