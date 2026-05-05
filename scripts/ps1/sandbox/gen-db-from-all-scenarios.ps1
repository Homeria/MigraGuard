# Purpose: Generate SQLite sandboxes (.db) for ALL scenarios in the scenarios directory.
# Usage:   .\scripts\ps1\sandbox\gen-db-from-all-scenarios.ps1 [--force]
param([switch]$Force)

$Scenarios = Get-ChildItem "experiments/scenarios/*.yaml"
foreach ($s in $Scenarios) {
    Write-Host "🔄 Processing: $($s.Name)" -ForegroundColor Yellow
    & "$PSScriptRoot/gen-db-from-scenario.ps1" -Scenario $s.FullName -Force:$Force
}
Write-Host "✅ All sandboxes seeded in experiments/data/" -ForegroundColor Green
