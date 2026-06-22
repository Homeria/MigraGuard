param(
    [int]$DurationSeconds = 35,
    [int]$WarmupSeconds = 8,
    [int]$ExtraOrders = 1000000,
    [string]$Workload = "orders_mixed",
    [string]$RunTag = "forecast_time_compression_20260621"
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$resultRoot = Join-Path $repoRoot "experiments\benchmark\results\$RunTag"
$summaryPath = Join-Path $resultRoot "forecast_time_compression_summary.csv"
$runLogPath = Join-Path $resultRoot "run_summary.log"

New-Item -ItemType Directory -Force -Path $resultRoot | Out-Null

$timeCases = @(
    [pscustomobject]@{
        case_id = "recommended_low"
        hour = "02:00"
        target_tps = 10
        forecast_meaning = "forecast recommended low-load window"
    },
    [pscustomobject]@{
        case_id = "safe_alternative"
        hour = "04:00"
        target_tps = 200
        forecast_meaning = "low-load alternative window"
    },
    [pscustomobject]@{
        case_id = "unexpected_event"
        hour = "02:00 + event"
        target_tps = 1000
        forecast_meaning = "traffic spike injected into recommended hour"
    },
    [pscustomobject]@{
        case_id = "peak_danger"
        hour = "10:00"
        target_tps = 5000
        forecast_meaning = "daytime peak / forecast danger window"
    }
)

$ddlCases = @(
    [pscustomobject]@{
        ddl_id = "DDL-2"
        ddl_kind = "concurrent index"
        ddl_path = ".\experiments\ddl\026_safe_concurrent_orders_composite_index.sql"
        forecast_recommended_hour = "02:00"
        forecast_recommended_level = "Safe"
    },
    [pscustomobject]@{
        ddl_id = "DDL-3"
        ddl_kind = "standard index"
        ddl_path = ".\experiments\ddl\037_linter_warning_standard_orders_index.sql"
        forecast_recommended_hour = "02:00"
        forecast_recommended_level = "Warning"
    }
)

function Get-Clients([int]$TargetTps) {
    if ($TargetTps -ge 5000) { return 24 }
    if ($TargetTps -ge 1000) { return 8 }
    if ($TargetTps -ge 200) { return 4 }
    return 4
}

function Get-Threads([int]$Clients) {
    if ($Clients -ge 8) { return 4 }
    return 2
}

function Get-LatestRunDir([string]$Prefix) {
    $dirs = Get-ChildItem -Directory (Join-Path $repoRoot "experiments\benchmark\results") |
        Where-Object { $_.Name -like "$Prefix*" } |
        Sort-Object LastWriteTime -Descending
    if ($dirs.Count -eq 0) {
        throw "No result directory found for prefix: $Prefix"
    }
    return $dirs[0].FullName
}

function Get-DbMetricSummary([string]$RunDir) {
    $metricsPath = Join-Path $RunDir "db_metrics.csv"
    if (-not (Test-Path $metricsPath)) {
        return [pscustomobject]@{
            max_connections = $null
            max_active_connections = $null
            max_lock_waiters = $null
        }
    }
    $rows = Import-Csv $metricsPath
    [pscustomobject]@{
        max_connections = (($rows | ForEach-Object { [int]$_.connections }) | Measure-Object -Maximum).Maximum
        max_active_connections = (($rows | ForEach-Object { [int]$_.active_connections }) | Measure-Object -Maximum).Maximum
        max_lock_waiters = (($rows | ForEach-Object { [int]$_.lock_waiters }) | Measure-Object -Maximum).Maximum
    }
}

function Read-RunSummary([string]$Path, [string]$Kind, [pscustomobject]$TimeCase, [pscustomobject]$DdlCase) {
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
    $dbMetrics = Get-DbMetricSummary -RunDir $Path
    $ddlDuration = if ($metadata.PSObject.Properties.Name -contains "ddl_duration_ms") { $metadata.ddl_duration_ms } else { $null }
    $ddlExit = if ($metadata.PSObject.Properties.Name -contains "ddl_exit_code") { $metadata.ddl_exit_code } else { $null }

    [pscustomobject]@{
        kind = $Kind
        ddl_id = $DdlCase.ddl_id
        ddl_kind = $DdlCase.ddl_kind
        hour = $TimeCase.hour
        case_id = $TimeCase.case_id
        forecast_meaning = $TimeCase.forecast_meaning
        forecast_recommended_hour = $DdlCase.forecast_recommended_hour
        forecast_recommended_level = $DdlCase.forecast_recommended_level
        target_tps = $TimeCase.target_tps
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
        max_connections = $dbMetrics.max_connections
        max_active_connections = $dbMetrics.max_active_connections
        max_lock_waiters = $dbMetrics.max_lock_waiters
    }
}

Push-Location $repoRoot
try {
    $rows = @()
    "Started at $(Get-Date -Format o)" | Set-Content -Encoding UTF8 -Path $runLogPath
    "DurationSeconds: $DurationSeconds" | Add-Content -Encoding UTF8 -Path $runLogPath
    "WarmupSeconds: $WarmupSeconds" | Add-Content -Encoding UTF8 -Path $runLogPath
    "ExtraOrders: $ExtraOrders" | Add-Content -Encoding UTF8 -Path $runLogPath

    $baselineByCase = @{}
    foreach ($timeCase in $timeCases) {
        $clients = Get-Clients -TargetTps $timeCase.target_tps
        $threads = Get-Threads -Clients $clients
        "Running baseline case=$($timeCase.case_id) target_tps=$($timeCase.target_tps) clients=$clients threads=$threads" |
            Tee-Object -FilePath $runLogPath -Append
        & (Join-Path $PSScriptRoot "run-baseline.ps1") `
            -Workload $Workload -TargetTps $timeCase.target_tps -Clients $clients -Threads $threads `
            -DurationSeconds $DurationSeconds -ExtraOrders $ExtraOrders -RunTag "${RunTag}_baseline_$($timeCase.case_id)"
        $baselinePrefix = "${RunTag}_baseline_$($timeCase.case_id)_baseline_${Workload}_$($timeCase.target_tps)tps"
        $baselineDir = Get-LatestRunDir $baselinePrefix
        $placeholderDdl = [pscustomobject]@{
            ddl_id = "BASELINE"
            ddl_kind = "baseline"
            forecast_recommended_hour = ""
            forecast_recommended_level = ""
        }
        $baselineRow = Read-RunSummary -Path $baselineDir -Kind "baseline" -TimeCase $timeCase -DdlCase $placeholderDdl
        $baselineByCase[$timeCase.case_id] = $baselineRow
        $rows += $baselineRow
    }

    foreach ($ddlCase in $ddlCases) {
        foreach ($timeCase in $timeCases) {
            $clients = Get-Clients -TargetTps $timeCase.target_tps
            $threads = Get-Threads -Clients $clients
            "Running ddl=$($ddlCase.ddl_id) case=$($timeCase.case_id) target_tps=$($timeCase.target_tps) clients=$clients threads=$threads" |
                Tee-Object -FilePath $runLogPath -Append
            & (Join-Path $PSScriptRoot "run-ddl-case.ps1") `
                -DdlPath $ddlCase.ddl_path -Workload $Workload -TargetTps $timeCase.target_tps -Clients $clients -Threads $threads `
                -DurationSeconds $DurationSeconds -WarmupSeconds $WarmupSeconds -ExtraOrders $ExtraOrders `
                -RunTag "${RunTag}_$($ddlCase.ddl_id)_$($timeCase.case_id)"
            $ddlName = [IO.Path]::GetFileNameWithoutExtension((Resolve-Path $ddlCase.ddl_path).Path)
            $ddlPrefix = "${RunTag}_$($ddlCase.ddl_id)_$($timeCase.case_id)_${ddlName}_${Workload}_$($timeCase.target_tps)tps"
            $ddlDir = Get-LatestRunDir $ddlPrefix
            $rows += Read-RunSummary -Path $ddlDir -Kind "ddl" -TimeCase $timeCase -DdlCase $ddlCase
        }
    }

    $enriched = foreach ($row in $rows) {
        $baseline = $baselineByCase[$row.case_id]
        $p99Ratio = if ($null -ne $baseline -and [double]$baseline.latency_p99_ms -gt 0) {
            [math]::Round([double]$row.latency_p99_ms / [double]$baseline.latency_p99_ms, 4)
        } else { $null }
        $tpsRatio = if ($null -ne $baseline -and [double]$baseline.actual_tps -gt 0) {
            [math]::Round([double]$row.actual_tps / [double]$baseline.actual_tps, 4)
        } else { $null }
        $activeRatio = if ($null -ne $baseline -and $null -ne $baseline.max_active_connections -and [double]$baseline.max_active_connections -gt 0) {
            [math]::Round([double]$row.max_active_connections / [double]$baseline.max_active_connections, 4)
        } else { $null }
        $row | Add-Member -NotePropertyName baseline_p99_ms -NotePropertyValue $(if ($null -ne $baseline) { $baseline.latency_p99_ms } else { $null }) -Force
        $row | Add-Member -NotePropertyName p99_ratio_to_baseline -NotePropertyValue $p99Ratio -Force
        $row | Add-Member -NotePropertyName tps_ratio_to_baseline -NotePropertyValue $tpsRatio -Force
        $row | Add-Member -NotePropertyName active_connection_ratio_to_baseline -NotePropertyValue $activeRatio -Force
        $row
    }

    $enriched | Export-Csv -NoTypeInformation -Encoding UTF8 -Path $summaryPath
    "Summary: $summaryPath" | Add-Content -Encoding UTF8 -Path $runLogPath
    Write-Host "Forecast time compression validation completed: $summaryPath"
}
finally {
    Pop-Location
}
