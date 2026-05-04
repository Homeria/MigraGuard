# MigraGuard Batch Research Runner (Windows PowerShell)
# Executes all DDL cases against all scenarios and accumulates results to CSV

$DataDir = "experiments/data"
$ScenarioDir = "experiments/scenarios"
$DdlDir = "experiments/ddl"
$ReportFile = "experiments/reports/research_results.csv"

if (-not (Test-Path "experiments/reports")) { New-Item -ItemType Directory -Path "experiments/reports" }

Write-Host "🚀 Starting Massive Batch Research Analysis..." -ForegroundColor Cyan
Write-Host "📊 Target Report: $ReportFile"

# 1. Initialize CSV with Header
$firstDdl = Get-ChildItem -Path $DdlDir -Filter *.sql | Select-Object -First 1
$firstDb = Get-ChildItem -Path $DataDir -Filter *.db | Select-Object -First 1

if ($null -eq $firstDdl -or $null -eq $firstDb) {
    Write-Error "No DDL or DB files found. Please run seed_all first."
    exit 1
}

# Run once to get header and first data row (or we could try to just get header)
go run ./cmd/migraguard analyze $firstDdl.FullName --sandbox $firstDb.FullName --output csv > $ReportFile

# 2. Actual Loop (Skip the first one we already did to avoid duplicates, or just overwrite)
$dbs = Get-ChildItem -Path $DataDir -Filter *.db
$ddls = Get-ChildItem -Path $DdlDir -Filter *.sql
$count = 0

foreach ($db in $dbs) {
    foreach ($ddl in $ddls) {
        # Skip the very first combination if we want to be perfect, but append is fine
        $count++
        Write-Host "[$count] Analyzing $($ddl.Name) against $($db.Name)..."
        
        go run ./cmd/migraguard analyze $ddl.FullName --sandbox $db.FullName --output csv --no-header >> $ReportFile
    }
}

Write-Host "--------------------------------------------------"
Write-Host "✅ Research complete. $count cases processed." -ForegroundColor Green
Write-Host "📈 Data saved to $ReportFile"
