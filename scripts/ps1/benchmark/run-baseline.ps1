param(
    [ValidateSet("orders_mixed", "logs_write", "stock_update", "balance_update")]
    [string]$Workload = "orders_mixed",
    [int]$TargetTps = 50,
    [int]$Clients = 10,
    [int]$Threads = 2,
    [int]$DurationSeconds = 30,
    [string]$Database = "benchmark_run",
    [string]$RunTag = "",
    [int]$ExtraOrders = 0,
    [switch]$SkipReset
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$safeRunTag = $RunTag -replace '[^A-Za-z0-9_-]', '_'
$runPrefix = if ([string]::IsNullOrWhiteSpace($safeRunTag)) { "" } else { "${safeRunTag}_" }
$runName = "${runPrefix}baseline_${Workload}_${TargetTps}tps_$stamp"
$hostResultDir = Join-Path $repoRoot "experiments\benchmark\results\$runName"
$containerResultDir = "/work/results/$runName"
$metricsPath = Join-Path $hostResultDir "db_metrics.csv"

New-Item -ItemType Directory -Force -Path $hostResultDir | Out-Null

Push-Location $repoRoot
try {
    docker compose up -d db benchmark | Out-Host
    if ($LASTEXITCODE -ne 0) {
        throw "Starting benchmark containers failed with exit code $LASTEXITCODE."
    }
    if (-not $SkipReset) {
        & (Join-Path $PSScriptRoot "reset-benchmark-db.ps1") -Database $Database -ExtraOrders $ExtraOrders
    }

    $metadata = [ordered]@{
        run_name = $runName
        started_at = (Get-Date).ToString("o")
        database = $Database
        workload = $Workload
        target_tps = $TargetTps
        clients = $Clients
        threads = $Threads
        duration_seconds = $DurationSeconds
        extra_orders = $ExtraOrders
    }
    $metadata | ConvertTo-Json | Set-Content (Join-Path $hostResultDir "metadata.json") -Encoding utf8

    $collectorScript = Join-Path $PSScriptRoot "collect-db-metrics.ps1"
    $metricJob = Start-Job -ScriptBlock {
        param($Script, $Output, $Duration, $Db)
        & $Script -OutputPath $Output -DurationSeconds $Duration -Database $Db
    } -ArgumentList $collectorScript, $metricsPath, $DurationSeconds, $Database

    $pgbenchArgs = @(
        "compose", "exec", "-T", "-w", $containerResultDir, "benchmark",
        "pgbench", "-h", "db", "-U", "user", "-n",
        "-f", "/work/sql/$Workload.sql",
        "-c", $Clients, "-j", $Threads, "-T", $DurationSeconds,
        "-R", $TargetTps, "-P", "5", "-l", $Database
    )
    $previousErrorAction = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $pgbenchOutput = & docker @pgbenchArgs 2>&1
    $pgbenchExit = $LASTEXITCODE
    $ErrorActionPreference = $previousErrorAction
    $pgbenchOutput | Set-Content (Join-Path $hostResultDir "pgbench_summary.txt") -Encoding utf8

    Wait-Job $metricJob | Out-Null
    Receive-Job $metricJob -ErrorAction Stop | Out-Host
    Remove-Job $metricJob

    if ($pgbenchExit -ne 0) {
        throw "pgbench failed. See $hostResultDir\pgbench_summary.txt"
    }

    python (Join-Path $repoRoot "tools\visualization\benchmark\summarize_pgbench.py") $hostResultDir | Out-Host
    if ($LASTEXITCODE -ne 0) {
        throw "Summarizing pgbench logs failed with exit code $LASTEXITCODE."
    }
}
finally {
    Pop-Location
}

Write-Host "Baseline run completed: $hostResultDir"
