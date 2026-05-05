# Purpose: Generate a single SQLite sandbox (.db) from a YAML scenario file.
# Usage:   .\scripts\ps1\sandbox\gen-db-from-scenario.ps1 <scenario_name> [--force]
param([string]$Scenario, [switch]$Force)

$SPath = "experiments/scenarios/$Scenario"
if (-not (Test-Path $SPath)) { $SPath = $Scenario }

$args = @("run", "./cmd/migraguard", "simulate", "--scenario", $SPath)
if ($Force) { $args += "--force" }

go $args

$dbName = (Get-Content $SPath | Select-String -Pattern 'experiment_name:\s*(.*)' | ForEach-Object { $_.Matches.Groups[1].Value.Trim().Replace("'", "").Replace('"', '') })
if (Test-Path "$dbName.db") {
    if (-not (Test-Path "experiments/data")) { New-Item -ItemType Directory "experiments/data" | Out-Null }
    Move-Item "$dbName.db" "experiments/data/" -Force
    Write-Host "[OK] Database generated: experiments/data/$dbName.db" -ForegroundColor Green
}
