# Purpose: Analyze a single SQL DDL case against an offline sandbox DB.
# Usage:   .\scripts\ps1\analyze\run-analysis-from-sandbox.ps1 <db_name> <sql_file> [output_type] [report_path]
# [output_type]: csv, markdown, console (default)
param([string]$DB, [string]$SQL, [string]$Output = "console", [string]$ReportPath)

$DBPath = "experiments/data/$DB"
if (-not (Test-Path $DBPath)) { $DBPath = $DB }

$SQLPath = "experiments/ddl/$SQL"
if (-not (Test-Path $SQLPath)) { $SQLPath = $SQL }

$args = @("run", "./cmd/migraguard", "analyze", $SQLPath, "--sandbox", $DBPath)
if ($Output -ne "console") { $args += "--output"; $args += $Output }

if ($ReportPath) {
    go $args | Out-File -FilePath $ReportPath -Encoding ascii
    Write-Host "[OK] Report saved to $ReportPath" -ForegroundColor Green
} else {
    go $args
}
