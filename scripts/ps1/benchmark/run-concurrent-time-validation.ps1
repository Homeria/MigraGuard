param(
    [string]$DdlPath = ".\experiments\ddl\026_safe_concurrent_orders_composite_index.sql",
    [string]$Workload = "orders_mixed",
    [int[]]$TargetTpsList = @(10, 1000, 5000),
    [int]$DurationSeconds = 45,
    [int]$WarmupSeconds = 10,
    [int]$ExtraOrders = 1000000,
    [string]$RunTag = "concurrently_time_20260621"
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$resultRoot = Join-Path $repoRoot "experiments\benchmark\results\$RunTag"
$summaryPath = Join-Path $resultRoot "concurrently_time_summary.csv"
$runLogPath = Join-Path $resultRoot "run_summary.log"

New-Item -ItemType Directory -Force -Path $resultRoot | Out-Null

function Get-LatestRunDir([string]$Prefix) {
    $dirs = Get-ChildItem -Directory (Join-Path $repoRoot "experiments\benchmark\results") |
        Where-Object { $_.Name -like "$Prefix*" } |
        Sort-Object LastWriteTime -Descending
    if ($dirs.Count -eq 0) {
        throw "No result directory found for prefix: $Prefix"
    }
    return $dirs[0].FullName
}

function Read-RunSummary([string]$Path, [string]$Kind, [int]$TargetTps) {
    $metricsPath = Join-Path $Path "summary_metrics.json"
    $metadataPath = Join-Path $Path "metadata.json"
    if (-not (Test-Path $metricsPath)) {
        throw "Missing summary_metrics.json in $Path"
    }
    if (-not (Test-Path $metadataPath)) {
        throw "Missing metadata.json in $Path"
    }
    $metrics = Get-Content -Encoding UTF8 -Path $metricsPath | ConvertFrom-Json
    $metadata = Get-Content -Encoding UTF8 -Path $metadataPath | ConvertFrom-Json
    $ddlDuration = if ($metadata.PSObject.Properties.Name -contains "ddl_duration_ms") { $metadata.ddl_duration_ms } else { $null }
    $ddlExit = if ($metadata.PSObject.Properties.Name -contains "ddl_exit_code") { $metadata.ddl_exit_code } else { $null }
    [pscustomobject]@{
        kind = $Kind
        target_tps = $TargetTps
        run_dir = (Split-Path -Leaf $Path)
        actual_tps = $metrics.actual_tps
        latency_avg_ms = $metrics.average_latency_ms
        latency_p50_ms = $metrics.p50_latency_ms
        latency_p95_ms = $metrics.p95_latency_ms
        latency_p99_ms = $metrics.p99_latency_ms
        latency_max_ms = $metrics.max_latency_ms
        failed_transactions = $metrics.failed_transactions
        ddl_duration_ms = $ddlDuration
        ddl_exit_code = $ddlExit
    }
}

Push-Location $repoRoot
try {
    $rows = @()
    "Started at $(Get-Date -Format o)" | Set-Content -Encoding UTF8 -Path $runLogPath
    "DDL: $DdlPath" | Add-Content -Encoding UTF8 -Path $runLogPath
    "ExtraOrders: $ExtraOrders" | Add-Content -Encoding UTF8 -Path $runLogPath

    foreach ($targetTps in $TargetTpsList) {
        $clients = if ($targetTps -ge 5000) { 24 } elseif ($targetTps -ge 1000) { 8 } else { 4 }
        $threads = if ($clients -ge 8) { 4 } else { 2 }

        "Running baseline target_tps=$targetTps clients=$clients threads=$threads" | Tee-Object -FilePath $runLogPath -Append
        & (Join-Path $PSScriptRoot "run-baseline.ps1") `
            -Workload $Workload -TargetTps $targetTps -Clients $clients -Threads $threads `
            -DurationSeconds $DurationSeconds -ExtraOrders $ExtraOrders -RunTag "${RunTag}_baseline"
        $baselineDir = Get-LatestRunDir "${RunTag}_baseline_baseline_${Workload}_${targetTps}tps"
        $rows += Read-RunSummary -Path $baselineDir -Kind "baseline" -TargetTps $targetTps

        "Running concurrent DDL target_tps=$targetTps clients=$clients threads=$threads" | Tee-Object -FilePath $runLogPath -Append
        & (Join-Path $PSScriptRoot "run-ddl-case.ps1") `
            -DdlPath $DdlPath -Workload $Workload -TargetTps $targetTps -Clients $clients -Threads $threads `
            -DurationSeconds $DurationSeconds -WarmupSeconds $WarmupSeconds -ExtraOrders $ExtraOrders `
            -RunTag "${RunTag}_ddl"
        $ddlName = [IO.Path]::GetFileNameWithoutExtension((Resolve-Path $DdlPath).Path)
        $ddlDir = Get-LatestRunDir "${RunTag}_ddl_${ddlName}_${Workload}_${targetTps}tps"
        $rows += Read-RunSummary -Path $ddlDir -Kind "ddl_concurrently" -TargetTps $targetTps
    }

    $baselineByTps = @{}
    foreach ($row in $rows) {
        if ($row.kind -eq "baseline") {
            $baselineByTps[$row.target_tps] = $row
        }
    }

    $enriched = foreach ($row in $rows) {
        $baseline = $baselineByTps[$row.target_tps]
        $p99Ratio = if ($null -ne $baseline -and [double]$baseline.latency_p99_ms -gt 0) {
            [math]::Round([double]$row.latency_p99_ms / [double]$baseline.latency_p99_ms, 4)
        } else { $null }
        $tpsRatio = if ($null -ne $baseline -and [double]$baseline.actual_tps -gt 0) {
            [math]::Round([double]$row.actual_tps / [double]$baseline.actual_tps, 4)
        } else { $null }
        $row | Add-Member -NotePropertyName baseline_p99_ms -NotePropertyValue $(if ($null -ne $baseline) { $baseline.latency_p99_ms } else { $null }) -Force
        $row | Add-Member -NotePropertyName p99_ratio_to_baseline -NotePropertyValue $p99Ratio -Force
        $row | Add-Member -NotePropertyName tps_ratio_to_baseline -NotePropertyValue $tpsRatio -Force
        $row
    }

    $enriched | Export-Csv -NoTypeInformation -Encoding UTF8 -Path $summaryPath
    "Summary: $summaryPath" | Add-Content -Encoding UTF8 -Path $runLogPath
    Write-Host "Concurrent time validation completed: $summaryPath"
}
finally {
    Pop-Location
}
