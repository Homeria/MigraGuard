# MigraGuard Batch Seeder (Windows PowerShell)
# Seeds all scenarios found in experiments/scenarios into experiments/data

$DataDir = "experiments/data"
$ScenarioDir = "experiments/scenarios"

if (-not (Test-Path $DataDir)) { New-Item -ItemType Directory -Path $DataDir }

Write-Host "🚀 Starting Batch Seeding for all scenarios..." -ForegroundColor Cyan

$Scenarios = Get-ChildItem -Path $ScenarioDir -Filter *.yaml

foreach ($Scenario in $Scenarios) {
    Write-Host "--------------------------------------------------"
    Write-Host "Processing: $($Scenario.Name)"
    
    go run ./cmd/migraguard simulate --scenario $Scenario.FullName --force
    
    # Extract experiment_name to move the resulting DB
    $experimentLine = Get-Content $Scenario.FullName | Select-String "experiment_name"
    $dbName = ($experimentLine -split ":")[1].Trim().Replace('"', '').Replace("'", "")
    $dbFile = "$dbName.db"
    
    if (Test-Path $dbFile) {
        Move-Item -Path $dbFile -Destination $DataDir -Force
        Write-Host "[OK] Generated: $DataDir/$dbFile" -ForegroundColor Green
    }
}

Write-Host "--------------------------------------------------"
Write-Host "✅ Batch seeding complete. All databases are in $DataDir" -ForegroundColor Cyan
