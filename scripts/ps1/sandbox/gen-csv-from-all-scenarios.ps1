# Purpose: Generate CSV metrics for ALL scenarios in the scenarios directory.
# Usage:   .\scripts\ps1\sandbox\gen-csv-from-all-scenarios.ps1 [--force]
param([switch]$Force)

$Scenarios = Get-ChildItem "experiments/scenarios/*.yaml"
$OutputDir = "experiments/reports/metrics"
if (-not (Test-Path $OutputDir)) { New-Item -ItemType Directory $OutputDir | Out-Null }

foreach ($s in $Scenarios) {
    $OutPath = Join-Path $OutputDir "$($s.BaseName)_metrics.csv"
    Write-Host "🔄 Exporting CSV for: $($s.Name)" -ForegroundColor Yellow
    & "$PSScriptRoot/gen-csv-from-scenario.ps1" -Scenario $s.FullName -Output $OutPath -Force:$Force
}
Write-Host "✅ All metrics exported to $OutputDir" -ForegroundColor Green
