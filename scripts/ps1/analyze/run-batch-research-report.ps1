# Purpose: Run ALL SQL cases against ALL Sandbox DBs and consolidate findings into a master CSV.
# Usage:   .\scripts\ps1\analyze\run-batch-research-report.ps1
$Report = "experiments/reports/research_results.csv"
if (-not (Test-Path "experiments/reports")) { New-Item -ItemType Directory "experiments/reports" | Out-Null }

$DBs = Get-ChildItem "experiments/data/*.db"
$SQLs = Get-ChildItem "experiments/ddl/*.sql"

if ($DBs.Count -eq 0 -or $SQLs.Count -eq 0) {
    Write-Error "Required files missing. Please run gen-db-from-all-scenarios.ps1 first."
    exit 1
}

Write-Host "🚀 Starting Massive Batch Research Analysis..." -ForegroundColor Cyan

$count = 0
foreach ($db in $DBs) {
    foreach ($sql in $SQLs) {
        $count++
        $headerFlag = if ($count -eq 1) { "" } else { "--no-header" }
        Write-Host "[$count] $($sql.Name) @ $($db.Name)" -ForegroundColor Gray
        
        go run ./cmd/migraguard analyze $sql.FullName --sandbox $db.FullName --output csv $headerFlag | Out-File -FilePath $Report -Encoding ascii -Append
    }
}
Write-Host "✅ Research complete. Master report: $Report" -ForegroundColor Green
