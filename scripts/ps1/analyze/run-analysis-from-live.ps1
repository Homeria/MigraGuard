# Purpose: Analyze a SQL DDL case against a LIVE database with local collector.
# Usage:   .\scripts\ps1\analyze\run-analysis-from-live.ps1 <sql_file> [db_url] [sqlite_path] [output_type]
param([string]$SQL, [string]$DBUrl, [string]$SQLite, [string]$Output = "console")

$SQLPath = "experiments/ddl/$SQL"
if (-not (Test-Path $SQLPath)) { $SQLPath = $SQL }

$args = @("run", "./cmd/migraguard", "analyze", $SQLPath)
if ($DBUrl) { $args += "--db"; $args += $DBUrl }
if ($SQLite) { $args += "--sqlite"; $args += $SQLite }
if ($Output -ne "console") { $args += "--output"; $args += $Output }

go $args
