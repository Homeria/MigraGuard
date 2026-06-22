param(
    [Parameter(Mandatory = $true)]
    [string]$OutputPath,
    [int]$DurationSeconds = 30,
    [string]$Database = "benchmark_run"
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$queryPath = "/work/sql/collect_metrics.sql"
$lines = [System.Collections.Generic.List[string]]::new()
$lines.Add("measured_at,connections,active_connections,lock_waiters,xact_commit,xact_rollback,deadlocks")

Push-Location $repoRoot
try {
    for ($second = 0; $second -lt $DurationSeconds; $second++) {
        $row = docker compose exec -T benchmark psql `
            -h db -U user -d $Database -At -F ',' -f $queryPath
        if ($LASTEXITCODE -ne 0) {
            throw "Metric collection failed at second $second."
        }
        $lines.Add(($row | Select-Object -Last 1))
        Start-Sleep -Seconds 1
    }
}
finally {
    Pop-Location
}

$parent = Split-Path $OutputPath -Parent
if ($parent) {
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
}
$lines | Set-Content -Path $OutputPath -Encoding utf8
