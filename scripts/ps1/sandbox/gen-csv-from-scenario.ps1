# Purpose: Generate a single CSV metrics file from a scenario.
# Usage:   .\scripts\ps1\sandbox\gen-csv-from-scenario.ps1 <scenario_name> [output_path] [--force]
param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$Scenario, 
    [Parameter(Position=1)]
    [string]$Output, 
    [switch]$Force
)

$SPath = "experiments/scenarios/$Scenario"
if (-not (Test-Path $SPath)) { $SPath = $Scenario }
if (-not (Test-Path $SPath)) { Write-Error "Scenario not found: $Scenario"; exit 1 }

# Auto-determine output path if not provided
if (-not $Output) {
    $base = [System.IO.Path]::GetFileNameWithoutExtension($SPath)
    $OutputDir = "experiments/reports/metrics"
    if (-not (Test-Path $OutputDir)) { New-Item -ItemType Directory $OutputDir | Out-Null }
    $Output = Join-Path $OutputDir "$($base)_metrics.csv"
}

Write-Host "📊 Generating Metrics CSV from $Scenario..." -ForegroundColor Yellow

$args = @("run", "./cmd/migraguard", "simulate", "--scenario", $SPath, "--csv", $Output, "--no-db")
if ($Force) { $args += "--force" }

go $args

if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Metrics CSV successfully generated at: $Output" -ForegroundColor Green
}
