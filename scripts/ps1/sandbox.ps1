# MigraGuard One-Click Sandbox Script (Windows PowerShell)
# Usage: .\scripts\sandbox.ps1 <scenario_name> <ddl_name> [-Docker] [-Force]

param (
    [Parameter(Mandatory=$true, Position=0)]
    [string]$ScenarioIn,
    
    [Parameter(Mandatory=$true, Position=1)]
    [string]$SQLIn,
    
    [switch]$Docker,
    [switch]$Force
)

# Resolve Paths
$Scenario = Join-Path "experiments/scenarios" $ScenarioIn
if (-not (Test-Path $Scenario)) { $Scenario = $ScenarioIn }

$SQL = Join-Path "experiments/ddl" $SQLIn
if (-not (Test-Path $SQL)) { $SQL = $SQLIn }

# 1. Extract Experiment Name from YAML
$experimentLine = Get-Content $Scenario | Select-String "experiment_name"
$dbName = ($experimentLine -split ":")[1].Trim().Replace('"', '').Replace("'", "")
$dbPath = "experiments/data/$dbName.db"

Write-Host "--------------------------------------------------" -ForegroundColor Cyan
Write-Host "🚀 MigraGuard Sandbox Runner (PowerShell)"
Write-Host "Scenario: $Scenario"
Write-Host "SQL:      $SQL"
Write-Host "Sandbox:  $dbPath"
Write-Host "--------------------------------------------------" -ForegroundColor Cyan

if ($Docker) {
    $forceFlag = if ($Force) { "--force" } else { "" }
    docker compose run --rm analyze-shell sh -c "go run ./cmd/migraguard simulate --scenario $Scenario $forceFlag && go run ./cmd/migraguard analyze $SQL --sandbox $dbPath"
} else {
    $simArgs = @("run", "./cmd/migraguard", "simulate", "--scenario", $Scenario)
    if ($Force) { $simArgs += "--force" }
    
    go $simArgs
    if ($LASTEXITCODE -eq 0) {
        go run ./cmd/migraguard analyze $SQL --sandbox $dbPath
    }
}
