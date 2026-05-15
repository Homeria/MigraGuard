# 🛡️ MigraGuard: Predictive 24-Hour Forecast Workflow
# Usage: .\scripts\ps1\analyze\run-predictive-forecast.ps1 <db_path> <ddl_path>

param (
    [Parameter(Mandatory=$true)]
    [string]$dbPath,
    [Parameter(Mandatory=$true)]
    [string]$ddlPath
)

$csvName = "predictive_forecast.csv"
$imgName = "predictive_risk_heatmap.png"

Write-Host "--------------------------------------------------------" -ForegroundColor Cyan
Write-Host "🚀 [1/2] Running Predictive Risk Analysis (Go Engine)..." -ForegroundColor Cyan
Write-Host "--------------------------------------------------------" -ForegroundColor Cyan

go run ./cmd/migraguard analyze "$ddlPath" --sandbox "$dbPath" --forecast --output console

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ [ERROR] Analysis failed. Skipping visualization." -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host ""
Write-Host "--------------------------------------------------------" -ForegroundColor Cyan
Write-Host "📊 [2/2] Generating Risk Heatmap Visualization (Python)..." -ForegroundColor Cyan
Write-Host "--------------------------------------------------------" -ForegroundColor Cyan

if (!(Test-Path $csvName)) {
    Write-Host "❌ [ERROR] Forecast CSV ($csvName) was not generated." -ForegroundColor Red
    exit 1
}

python ./tools/visualization/analyze/plot_predictive_heatmap.py $csvName

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ [SUCCESS] Analysis complete!" -ForegroundColor Green
    Write-Host "📍 Report Location: $imgName" -ForegroundColor Green
} else {
    Write-Host "❌ [ERROR] Visualization failed." -ForegroundColor Red
    exit 1
}
