param(
    [Parameter(Mandatory = $true)]
    [string]$DdlPath,
    [ValidateSet("orders_mixed", "logs_write", "stock_update", "balance_update")]
    [string]$Workload = "orders_mixed",
    [int]$TargetTps = 50,
    [int]$Clients = 10,
    [int]$Threads = 2,
    [int]$DurationSeconds = 30,
    [int]$WarmupSeconds = 5,
    [string]$Database = "benchmark_run",
    [string]$MigraGuardBinary = ".\bin\migraguard.exe",
    [string]$MigraGuardConfig = ".\experiments\configs\cases\benchmark_calibrated.yaml",
    [string]$RunTag = "",
    [int]$ExtraOrders = 0,
    [string]$PreSqlPath = "",
    [switch]$SkipMigraGuard,
    [switch]$SkipReset,
    [switch]$AllowDdlFailure
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$resolvedDdl = (Resolve-Path $DdlPath).Path
$ddlName = [IO.Path]::GetFileNameWithoutExtension($resolvedDdl)
$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$safeRunTag = $RunTag -replace '[^A-Za-z0-9_-]', '_'
$runPrefix = if ([string]::IsNullOrWhiteSpace($safeRunTag)) { "" } else { "${safeRunTag}_" }
$runName = "${runPrefix}${ddlName}_${Workload}_${TargetTps}tps_$stamp"
$hostResultDir = Join-Path $repoRoot "experiments\benchmark\results\$runName"
$containerResultDir = "/work/results/$runName"
$metricsPath = Join-Path $hostResultDir "db_metrics.csv"
$metricJob = $null
$benchmarkJob = $null
$agentProcess = $null

if ($WarmupSeconds -ge $DurationSeconds) {
    throw "WarmupSeconds must be smaller than DurationSeconds."
}
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

    if (-not [string]::IsNullOrWhiteSpace($PreSqlPath)) {
        $resolvedPreSql = (Resolve-Path $PreSqlPath).Path
        $preSqlOutput = Get-Content $resolvedPreSql -Raw | docker compose exec -T db psql -U user -d $Database -v ON_ERROR_STOP=1 2>&1
        $preSqlExit = $LASTEXITCODE
        $preSqlOutput | Set-Content (Join-Path $hostResultDir "pre_sql_execution.txt") -Encoding utf8
        if ($preSqlExit -ne 0) {
            throw "PreSql execution failed. See $hostResultDir\pre_sql_execution.txt"
        }
    }

    $sqlitePath = Join-Path $hostResultDir "migraguard_metrics.db"
    $mgBinaryPath = Join-Path $repoRoot ($MigraGuardBinary -replace '^\.\\', '')
    $mgConfigPath = Join-Path $repoRoot ($MigraGuardConfig -replace '^\.\\', '')
    $dbUrl = "postgres://user:pass@127.0.0.1:55432/$Database`?sslmode=disable"

    if (-not $SkipMigraGuard) {
        if (-not (Test-Path $mgBinaryPath)) {
            throw "MigraGuard binary not found: $mgBinaryPath. Run 'go build -o bin/migraguard.exe ./cmd/migraguard'."
        }
        $agentStdout = Join-Path $hostResultDir "migraguard_agent.log"
        $agentStderr = Join-Path $hostResultDir "migraguard_agent_error.log"
        $agentArgs = @(
            "--config", $mgConfigPath, "agent",
            "--db", $dbUrl, "--sqlite", $sqlitePath,
            "--interval", "1", "--tables", "orders"
        )
        $agentProcess = Start-Process -FilePath $mgBinaryPath -ArgumentList $agentArgs `
            -RedirectStandardOutput $agentStdout -RedirectStandardError $agentStderr `
            -WindowStyle Hidden -PassThru
    }

    $previousErrorAction = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $squawkOutput = & "$env:ProgramFiles\nodejs\npx.cmd" --yes squawk-cli@2.58.0 $resolvedDdl 2>&1
    $squawkExit = $LASTEXITCODE
    $ErrorActionPreference = $previousErrorAction
    $squawkOutput | Set-Content (Join-Path $hostResultDir "squawk.txt") -Encoding utf8

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
    $benchmarkJob = Start-Job -ScriptBlock {
        param($Root, $Arguments)
        Push-Location $Root
        try {
            $ErrorActionPreference = "Continue"
            $output = & docker @Arguments 2>&1
            [pscustomobject]@{ ExitCode = $LASTEXITCODE; Output = @($output) }
        }
        finally {
            Pop-Location
        }
    } -ArgumentList $repoRoot, (, $pgbenchArgs)

    Start-Sleep -Seconds $WarmupSeconds

    $migraGuardExit = $null
    if (-not $SkipMigraGuard) {
        if ($null -ne $agentProcess -and -not $agentProcess.HasExited) {
            Stop-Process -Id $agentProcess.Id -Force -ErrorAction SilentlyContinue
            $agentProcess.WaitForExit()
        }
        $predictionCsv = Join-Path $hostResultDir "migraguard_prediction.csv"
        $predictionError = Join-Path $hostResultDir "migraguard_prediction_error.log"
        $analyzeArgs = @(
            "--config", $mgConfigPath, "analyze", $resolvedDdl,
            "--db", $dbUrl, "--sqlite", $sqlitePath,
            "--output", "csv"
        )
        $analyzeProcess = Start-Process -FilePath $mgBinaryPath -ArgumentList $analyzeArgs `
            -RedirectStandardOutput $predictionCsv -RedirectStandardError $predictionError `
            -WindowStyle Hidden -Wait -PassThru
        $migraGuardExit = $analyzeProcess.ExitCode
    }

    $ddlStopwatch = [Diagnostics.Stopwatch]::StartNew()
    $ErrorActionPreference = "Continue"
    $ddlOutput = Get-Content $resolvedDdl -Raw | docker compose exec -T db psql -U user -d $Database -v ON_ERROR_STOP=1 2>&1
    $ddlExit = $LASTEXITCODE
    $ErrorActionPreference = $previousErrorAction
    $ddlStopwatch.Stop()
    $ddlOutput | Set-Content (Join-Path $hostResultDir "ddl_execution.txt") -Encoding utf8

    Wait-Job $benchmarkJob | Out-Null
    $benchmarkResult = Receive-Job $benchmarkJob -ErrorAction Stop
    $benchmarkResult.Output | Set-Content (Join-Path $hostResultDir "pgbench_summary.txt") -Encoding utf8
    if ($benchmarkResult.ExitCode -ne 0) {
        throw "pgbench failed with exit code $($benchmarkResult.ExitCode)."
    }

    Wait-Job $metricJob | Out-Null
    Receive-Job $metricJob -ErrorAction Stop | Out-Host

    $metadata = [ordered]@{
        run_name = $runName
        started_at = (Get-Date).AddSeconds(-$DurationSeconds).ToString("o")
        ddl = $resolvedDdl
        ddl_exit_code = $ddlExit
        ddl_duration_ms = [math]::Round($ddlStopwatch.Elapsed.TotalMilliseconds, 3)
        squawk_exit_code = $squawkExit
        migraguard_exit_code = $migraGuardExit
        database = $Database
        workload = $Workload
        target_tps = $TargetTps
        clients = $Clients
        threads = $Threads
        duration_seconds = $DurationSeconds
        warmup_seconds = $WarmupSeconds
        extra_orders = $ExtraOrders
        pre_sql = $PreSqlPath
    }
    $metadata | ConvertTo-Json | Set-Content (Join-Path $hostResultDir "metadata.json") -Encoding utf8

    python (Join-Path $repoRoot "tools\visualization\benchmark\summarize_pgbench.py") $hostResultDir | Out-Host
    if ($LASTEXITCODE -ne 0) {
        throw "Summarizing pgbench logs failed with exit code $LASTEXITCODE."
    }
    if ($ddlExit -ne 0 -and -not $AllowDdlFailure) {
        throw "DDL execution failed. See $hostResultDir\ddl_execution.txt"
    }
}
finally {
    foreach ($job in @($benchmarkJob, $metricJob)) {
        if ($null -ne $job) {
            Remove-Job $job -Force -ErrorAction SilentlyContinue
        }
    }
    if ($null -ne $agentProcess -and -not $agentProcess.HasExited) {
        Stop-Process -Id $agentProcess.Id -Force -ErrorAction SilentlyContinue
    }
    Pop-Location
}

Write-Host "DDL benchmark completed: $hostResultDir"
