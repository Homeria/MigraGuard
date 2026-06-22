param(
    [int]$DurationSeconds = 20,
    [int]$WarmupSeconds = 5,
    [int]$ExtraOrders = 1000000,
    [string]$Workload = "orders_mixed",
    [string]$SandboxDb = ".\exp_21_time_recommendation_orders.db",
    [string]$RunTag = "forecast_hourly_optimality_20260621",
    [int]$MinTargetTps = 10
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$resultRoot = Join-Path $repoRoot "experiments\benchmark\results\$RunTag"
$forecastRoot = Join-Path $resultRoot "forecast"
$summaryPath = Join-Path $resultRoot "hourly_optimality_summary.csv"
$rankPath = Join-Path $resultRoot "hourly_optimality_operation_ranking.csv"
$runLogPath = Join-Path $resultRoot "run_summary.log"

New-Item -ItemType Directory -Force -Path $resultRoot | Out-Null
New-Item -ItemType Directory -Force -Path $forecastRoot | Out-Null

$ddlCases = @(
    [pscustomobject]@{
        ddl_id = "ONLINE-1"
        ddl_kind = "CREATE INDEX CONCURRENTLY"
        ddl_path = ".\experiments\ddl\026_safe_concurrent_orders_composite_index.sql"
        pre_sql_path = ""
    },
    [pscustomobject]@{
        ddl_id = "ONLINE-2"
        ddl_kind = "CREATE UNIQUE INDEX CONCURRENTLY"
        ddl_path = ".\experiments\ddl\047_online_unique_index_orders_order_no_concurrently.sql"
        pre_sql_path = ""
    },
    [pscustomobject]@{
        ddl_id = "ONLINE-3"
        ddl_kind = "ADD CONSTRAINT NOT VALID"
        ddl_path = ".\experiments\ddl\044_online_add_orders_status_check_not_valid.sql"
        pre_sql_path = ""
    },
    [pscustomobject]@{
        ddl_id = "ONLINE-4"
        ddl_kind = "VALIDATE CONSTRAINT"
        ddl_path = ".\experiments\ddl\045_online_validate_orders_status_check.sql"
        pre_sql_path = ".\experiments\ddl\045_setup_orders_status_check_not_valid.sql"
    }
)

function Get-Clients([int]$TargetTps) {
    if ($TargetTps -ge 4000) { return 24 }
    if ($TargetTps -ge 2000) { return 16 }
    if ($TargetTps -ge 1000) { return 8 }
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

function Find-LatestRunDir([string]$Prefix) {
    $dirs = Get-ChildItem -Directory (Join-Path $repoRoot "experiments\benchmark\results") |
        Where-Object { $_.Name -like "$Prefix*" -and (Test-Path (Join-Path $_.FullName "summary_metrics.json")) } |
        Sort-Object LastWriteTime -Descending
    if ($dirs.Count -eq 0) {
        return $null
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

function Read-RunSummary([string]$Path, [string]$Kind, [pscustomobject]$HourCase, [pscustomobject]$DdlCase) {
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
        hour = $HourCase.hour
        expected_tps = $HourCase.expected_tps
        target_tps = $HourCase.target_tps
        forecast_risk_score = $HourCase.forecast_risk_score
        forecast_risk_level = $HourCase.forecast_risk_level
        forecast_is_best_hour = $HourCase.forecast_is_best_hour
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

function Run-Forecast([pscustomobject]$DdlCase) {
    $ddlName = [IO.Path]::GetFileNameWithoutExtension((Resolve-Path $DdlCase.ddl_path).Path)
    $outDir = Join-Path $forecastRoot "$($DdlCase.ddl_id)_$ddlName"
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null

    Push-Location $repoRoot
    try {
        Remove-Item "predictive_forecast.csv" -ErrorAction SilentlyContinue
        $forecastLog = Join-Path $outDir "forecast_console.log"
        $previousErrorAction = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        & ".\bin\migraguard.exe" analyze $DdlCase.ddl_path --sandbox $SandboxDb --forecast --output console *>&1 |
            Set-Content -Encoding UTF8 -Path $forecastLog
        $forecastExit = $LASTEXITCODE
        $ErrorActionPreference = $previousErrorAction
        if (-not (Test-Path "predictive_forecast.csv")) {
            throw "Forecast CSV was not generated for $($DdlCase.ddl_id)."
        }
        if ($forecastExit -ne 0) {
            "Forecast command returned exit code $forecastExit, but predictive_forecast.csv was generated. Continuing." |
                Add-Content -Encoding UTF8 -Path $forecastLog
        }
        Move-Item -Force "predictive_forecast.csv" (Join-Path $outDir "predictive_forecast.csv")
    }
    finally {
        Pop-Location
    }

    return Import-Csv (Join-Path $outDir "predictive_forecast.csv") |
        Sort-Object { [int]$_.Hour } |
        ForEach-Object {
            $expectedTps = [double]$_.ExpectedTPS
            $targetTps = [int][math]::Round([math]::Max($MinTargetTps, $expectedTps))
            [pscustomobject]@{
                hour = ("{0:D2}:00" -f [int]$_.Hour)
                hour_number = [int]$_.Hour
                expected_tps = [math]::Round($expectedTps, 2)
                target_tps = $targetTps
                forecast_risk_score = [double]$_.RiskScore
                forecast_risk_level = $_.RiskLevel
                forecast_is_best_hour = ([string]$_.IsBestHour).ToLowerInvariant() -eq "true"
            }
        }
}

function Write-CurrentOutputs([object[]]$Rows) {
    if ($Rows.Count -eq 0) {
        return
    }

    $baselineByHour = @{}
    foreach ($row in $Rows) {
        if ($row.kind -eq "baseline") {
            $baselineByHour[$row.hour] = $row
        }
    }

    $enriched = foreach ($row in $Rows) {
        $baseline = $baselineByHour[$row.hour]
        $p99Ratio = if ($null -ne $baseline -and [double]$baseline.latency_p99_ms -gt 0) {
            [math]::Round([double]$row.latency_p99_ms / [double]$baseline.latency_p99_ms, 4)
        } else { $null }
        $tpsRatio = if ($null -ne $baseline -and [double]$baseline.actual_tps -gt 0) {
            [math]::Round([double]$row.actual_tps / [double]$baseline.actual_tps, 4)
        } else { $null }
        $activeRatio = if ($null -ne $baseline -and $null -ne $baseline.max_active_connections -and [double]$baseline.max_active_connections -gt 0) {
            [math]::Round([double]$row.max_active_connections / [double]$baseline.max_active_connections, 4)
        } else { $null }
        $impactScore = $null
        if ($row.kind -eq "ddl" -and $null -ne $p99Ratio) {
            $absoluteP99 = if ($null -ne $row.latency_p99_ms -and $row.latency_p99_ms -ne "") { [double]$row.latency_p99_ms } else { 0.0 }
            $tpsDropPenalty = [math]::Max(0.0, 1.0 - [double]$tpsRatio) * 50.0
            $ddlSeconds = if ($null -ne $row.ddl_duration_ms -and $row.ddl_duration_ms -ne "") { [double]$row.ddl_duration_ms / 1000.0 } else { 0.0 }
            $failurePenalty = if ($null -ne $row.failed_transactions -and $row.failed_transactions -ne "") { [double]$row.failed_transactions * 100.0 } else { 0.0 }
            $lockPenalty = if ($null -ne $row.max_lock_waiters -and $row.max_lock_waiters -ne "") { [double]$row.max_lock_waiters * 5.0 } else { 0.0 }
            $impactScore = [math]::Round($absoluteP99 + ([double]$p99Ratio * 2.0) + $tpsDropPenalty + $ddlSeconds + $failurePenalty + $lockPenalty, 4)
        }

        $row | Add-Member -NotePropertyName baseline_p99_ms -NotePropertyValue $(if ($null -ne $baseline) { $baseline.latency_p99_ms } else { $null }) -Force
        $row | Add-Member -NotePropertyName p99_ratio_to_baseline -NotePropertyValue $p99Ratio -Force
        $row | Add-Member -NotePropertyName tps_ratio_to_baseline -NotePropertyValue $tpsRatio -Force
        $row | Add-Member -NotePropertyName active_connection_ratio_to_baseline -NotePropertyValue $activeRatio -Force
        $row | Add-Member -NotePropertyName actual_operation_score -NotePropertyValue $impactScore -Force
        $row
    }

    $enriched | Export-Csv -NoTypeInformation -Encoding UTF8 -Path $summaryPath

    $ranked = @()
    foreach ($group in ($enriched | Where-Object { $_.kind -eq "ddl" -and $null -ne $_.actual_operation_score } | Group-Object ddl_id)) {
        $sorted = @($group.Group | Sort-Object {[double]$_.actual_operation_score}, {[int]($_.hour.Substring(0, 2))})
        for ($i = 0; $i -lt $sorted.Count; $i++) {
            $item = $sorted[$i]
            $item | Add-Member -NotePropertyName actual_operation_rank -NotePropertyValue ($i + 1) -Force
            $item | Add-Member -NotePropertyName is_operation_best -NotePropertyValue ($i -eq 0) -Force
            $item | Add-Member -NotePropertyName is_operation_top3 -NotePropertyValue ($i -lt 3) -Force
            $ranked += $item
        }
    }
    $ranked | Export-Csv -NoTypeInformation -Encoding UTF8 -Path $rankPath
}

Push-Location $repoRoot
try {
    $rows = @()
    "Started at $(Get-Date -Format o)" | Set-Content -Encoding UTF8 -Path $runLogPath
    "DurationSeconds: $DurationSeconds" | Add-Content -Encoding UTF8 -Path $runLogPath
    "WarmupSeconds: $WarmupSeconds" | Add-Content -Encoding UTF8 -Path $runLogPath
    "ExtraOrders: $ExtraOrders" | Add-Content -Encoding UTF8 -Path $runLogPath
    "MinTargetTps: $MinTargetTps" | Add-Content -Encoding UTF8 -Path $runLogPath

    $forecastByDdl = @{}
    foreach ($ddlCase in $ddlCases) {
        "Running forecast ddl=$($ddlCase.ddl_id) $($ddlCase.ddl_kind)" | Tee-Object -FilePath $runLogPath -Append
        $forecastByDdl[$ddlCase.ddl_id] = @(Run-Forecast -DdlCase $ddlCase)
    }

    $baselineHours = @($forecastByDdl[$ddlCases[0].ddl_id])
    $placeholderDdl = [pscustomobject]@{
        ddl_id = "BASELINE"
        ddl_kind = "baseline"
    }

    foreach ($hourCase in $baselineHours) {
        $clients = Get-Clients -TargetTps $hourCase.target_tps
        $threads = Get-Threads -Clients $clients
        $baselinePrefix = "${RunTag}_baseline_h$($hourCase.hour_number)_baseline_${Workload}_$($hourCase.target_tps)tps"
        $baselineDir = Find-LatestRunDir $baselinePrefix
        if ($null -eq $baselineDir) {
            "Running baseline hour=$($hourCase.hour) target_tps=$($hourCase.target_tps) clients=$clients threads=$threads" |
                Tee-Object -FilePath $runLogPath -Append
            & (Join-Path $PSScriptRoot "run-baseline.ps1") `
                -Workload $Workload -TargetTps $hourCase.target_tps -Clients $clients -Threads $threads `
                -DurationSeconds $DurationSeconds -ExtraOrders $ExtraOrders -RunTag "${RunTag}_baseline_h$($hourCase.hour_number)"
            $baselineDir = Get-LatestRunDir $baselinePrefix
        } else {
            "Reusing baseline hour=$($hourCase.hour) target_tps=$($hourCase.target_tps): $baselineDir" |
                Tee-Object -FilePath $runLogPath -Append
        }
        $rows += Read-RunSummary -Path $baselineDir -Kind "baseline" -HourCase $hourCase -DdlCase $placeholderDdl
        Write-CurrentOutputs -Rows $rows
    }

    foreach ($ddlCase in $ddlCases) {
        foreach ($hourCase in $forecastByDdl[$ddlCase.ddl_id]) {
            $clients = Get-Clients -TargetTps $hourCase.target_tps
            $threads = Get-Threads -Clients $clients
            $ddlName = [IO.Path]::GetFileNameWithoutExtension((Resolve-Path $ddlCase.ddl_path).Path)
            $ddlPrefix = "${RunTag}_$($ddlCase.ddl_id)_h$($hourCase.hour_number)_${ddlName}_${Workload}_$($hourCase.target_tps)tps"
            $ddlDir = Find-LatestRunDir $ddlPrefix
            if ($null -eq $ddlDir) {
                "Running ddl=$($ddlCase.ddl_id) hour=$($hourCase.hour) target_tps=$($hourCase.target_tps) clients=$clients threads=$threads" |
                    Tee-Object -FilePath $runLogPath -Append
                if ([string]::IsNullOrWhiteSpace($ddlCase.pre_sql_path)) {
                    & (Join-Path $PSScriptRoot "run-ddl-case.ps1") `
                        -DdlPath $ddlCase.ddl_path -Workload $Workload -TargetTps $hourCase.target_tps `
                        -Clients $clients -Threads $threads -DurationSeconds $DurationSeconds `
                        -WarmupSeconds $WarmupSeconds -ExtraOrders $ExtraOrders `
                        -RunTag "${RunTag}_$($ddlCase.ddl_id)_h$($hourCase.hour_number)"
                } else {
                    & (Join-Path $PSScriptRoot "run-ddl-case.ps1") `
                        -DdlPath $ddlCase.ddl_path -Workload $Workload -TargetTps $hourCase.target_tps `
                        -Clients $clients -Threads $threads -DurationSeconds $DurationSeconds `
                        -WarmupSeconds $WarmupSeconds -ExtraOrders $ExtraOrders `
                        -RunTag "${RunTag}_$($ddlCase.ddl_id)_h$($hourCase.hour_number)" `
                        -PreSqlPath $ddlCase.pre_sql_path
                }
                $ddlDir = Get-LatestRunDir $ddlPrefix
            } else {
                "Reusing ddl=$($ddlCase.ddl_id) hour=$($hourCase.hour) target_tps=$($hourCase.target_tps): $ddlDir" |
                    Tee-Object -FilePath $runLogPath -Append
            }
            $rows += Read-RunSummary -Path $ddlDir -Kind "ddl" -HourCase $hourCase -DdlCase $ddlCase
            Write-CurrentOutputs -Rows $rows
        }
    }

    Write-CurrentOutputs -Rows $rows
    "Summary: $summaryPath" | Add-Content -Encoding UTF8 -Path $runLogPath
    "Ranking: $rankPath" | Add-Content -Encoding UTF8 -Path $runLogPath
    Write-Host "Forecast hourly optimality validation completed: $summaryPath"
}
finally {
    Pop-Location
}
