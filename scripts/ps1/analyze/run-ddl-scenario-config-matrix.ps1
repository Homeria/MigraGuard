# Purpose: Run every DDL against every simulation scenario and every case config.
# Output:
#   - experiments/reports/batch_runs/<timestamp>/matrix_results.csv
#   - experiments/reports/batch_runs/<timestamp>/ddl_summary.csv
#   - experiments/reports/batch_runs/<timestamp>/ddl_by_config_summary.csv

param (
    [string]$RunName = "",
    [switch]$SkipBuild,
    [switch]$KeepSandboxDbs,
    [int]$DdlStart = 0,
    [int]$DdlEnd = 999,
    [string[]]$ConfigNames = @()
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ScriptDir\..\..\.."
Push-Location $ProjectRoot

try {
    $timestamp = if ([string]::IsNullOrWhiteSpace($RunName)) {
        Get-Date -Format "yyyyMMdd_HHmmss"
    } else {
        $RunName
    }

    $runDir = Join-Path $ProjectRoot "experiments\reports\batch_runs\$timestamp"
    $sandboxDir = Join-Path $runDir "sandboxes"
    New-Item -ItemType Directory -Force -Path $runDir | Out-Null
    New-Item -ItemType Directory -Force -Path $sandboxDir | Out-Null

    $binary = Join-Path $ProjectRoot "build\migraguard.exe"
    if (-not $SkipBuild) {
        New-Item -ItemType Directory -Force -Path (Split-Path $binary -Parent) | Out-Null
        Write-Host "[BUILD] $binary"
        & go build -o $binary ./cmd/migraguard
        if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code $LASTEXITCODE" }
    }

    if (-not (Test-Path $binary)) {
        $fallback = Join-Path $ProjectRoot "migraguard.exe"
        if (Test-Path $fallback) {
            $binary = $fallback
        } else {
            throw "MigraGuard binary not found. Run without -SkipBuild first."
        }
    }

    $scenarios = Get-ChildItem "experiments\scenarios" -File -Filter "*.yaml" | Sort-Object Name
    $ddls = @(Get-ChildItem "experiments\ddl" -File -Filter "*.sql" |
        Where-Object {
            $_.BaseName -match '^(\d{3})_' -and
            [int]$Matches[1] -ge $DdlStart -and
            [int]$Matches[1] -le $DdlEnd
        } |
        Sort-Object Name)
    $configs = @(Get-ChildItem "experiments\configs\cases" -File -Filter "*.yaml" |
        Where-Object {
            $ConfigNames.Count -eq 0 -or $_.BaseName -in $ConfigNames
        } |
        Sort-Object Name)

    if ($scenarios.Count -eq 0) { throw "No scenarios found." }
    if ($ddls.Count -eq 0) { throw "No DDL files found." }
    if ($configs.Count -eq 0) { throw "No config files found." }

    Write-Host "[INPUT] scenarios=$($scenarios.Count), ddls=$($ddls.Count), configs=$($configs.Count)"

    $scenarioDbMap = @{}
    foreach ($scenario in $scenarios) {
        Write-Host "[SIMULATE] $($scenario.Name)"
        $previousErrorActionPreference = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        & $binary simulate --scenario $scenario.FullName --force
        $simExit = $LASTEXITCODE
        $ErrorActionPreference = $previousErrorActionPreference
        if ($simExit -ne 0) {
            throw "Simulation failed for $($scenario.Name) with exit code $simExit."
        }

        $experimentLine = Select-String -Path $scenario.FullName -Pattern '^\s*experiment_name:\s*"?([^"]+)"?\s*$' | Select-Object -First 1
        if (-not $experimentLine) {
            throw "experiment_name not found in $($scenario.Name)."
        }
        $dbName = "$($experimentLine.Matches[0].Groups[1].Value).db"

        $rootDb = Join-Path $ProjectRoot $dbName
        if (-not (Test-Path $rootDb)) {
            throw "Could not find generated DB for $($scenario.Name)."
        }

        $destDb = Join-Path $sandboxDir (Split-Path $rootDb -Leaf)
        Move-Item -LiteralPath $rootDb -Destination $destDb -Force
        $scenarioDbMap[$scenario.Name] = $destDb
    }

    $results = New-Object System.Collections.Generic.List[object]
    $total = $scenarios.Count * $ddls.Count * $configs.Count
    $index = 0

    foreach ($scenario in $scenarios) {
        $dbPath = $scenarioDbMap[$scenario.Name]
        $scenarioName = [System.IO.Path]::GetFileNameWithoutExtension($scenario.Name)

        foreach ($config in $configs) {
            $configName = [System.IO.Path]::GetFileNameWithoutExtension($config.Name)

            foreach ($ddl in $ddls) {
                $index++
                $ddlName = [System.IO.Path]::GetFileNameWithoutExtension($ddl.Name)
                $intended = "unknown"
                if ($ddlName -match "_linter_warning_") { $intended = "LinterWarning" }
                elseif ($ddlName -match "_safe_") { $intended = "Safe" }
                elseif ($ddlName -match "_warning_") { $intended = "Warning" }
                elseif ($ddlName -match "_danger_") { $intended = "Danger" }

                Write-Progress -Activity "MigraGuard matrix analysis" -Status "$index / $total : $scenarioName | $configName | $ddlName" -PercentComplete (($index / $total) * 100)

                $previousErrorActionPreference = $ErrorActionPreference
                $ErrorActionPreference = "Continue"
                $csvText = & $binary analyze $ddl.FullName --sandbox $dbPath --config $config.FullName --output csv --no-header 2>$null
                $exitCode = $LASTEXITCODE
                $ErrorActionPreference = $previousErrorActionPreference

                if ([string]::IsNullOrWhiteSpace($csvText)) {
                    $results.Add([pscustomobject]@{
                        Scenario = $scenarioName
                        Config = $configName
                        DDL = $ddlName
                        IntendedLevel = $intended
                        ActualLevel = "ERROR"
                        RiskScore = ""
                        T_ddl = ""
                        T_block = ""
                        C_peak = ""
                        T_rec = ""
                        BaseTPS = ""
                        TPSSource = ""
                        TableName = ""
                        Operation = ""
                        LockLevel = ""
                        RewriteRequired = ""
                        TableSize = ""
                        ExitCode = $exitCode
                    })
                    continue
                }

                $row = $csvText | ConvertFrom-Csv -Header Timestamp,ScenarioCsv,SQLFile,TableName,Operation,LockLevel,RewriteRequired,RiskScore,RiskLevel,T_ddl,T_block,C_peak,T_rec,BaseTPS,TPSSource,TableSize

                $results.Add([pscustomobject]@{
                    Scenario = $scenarioName
                    Config = $configName
                    DDL = $ddlName
                    IntendedLevel = $intended
                    ActualLevel = $row.RiskLevel
                    RiskScore = $row.RiskScore
                    T_ddl = $row.T_ddl
                    T_block = $row.T_block
                    C_peak = $row.C_peak
                    T_rec = $row.T_rec
                    BaseTPS = $row.BaseTPS
                    TPSSource = $row.TPSSource
                    TableName = $row.TableName
                    Operation = $row.Operation
                    LockLevel = $row.LockLevel
                    RewriteRequired = $row.RewriteRequired
                    TableSize = $row.TableSize
                    ExitCode = $exitCode
                })
            }
        }
    }

    Write-Progress -Activity "MigraGuard matrix analysis" -Completed

    $matrixPath = Join-Path $runDir "matrix_results.csv"
    $results | Export-Csv -Path $matrixPath -NoTypeInformation -Encoding UTF8

    $ddlSummaryPath = Join-Path $runDir "ddl_summary.csv"
    $results |
        Group-Object DDL, IntendedLevel |
        ForEach-Object {
            $items = $_.Group
            [pscustomobject]@{
                DDL = $items[0].DDL
                IntendedLevel = $items[0].IntendedLevel
                Total = $items.Count
                Safe = @($items | Where-Object ActualLevel -eq "Safe").Count
                Warning = @($items | Where-Object ActualLevel -eq "Warning").Count
                Danger = @($items | Where-Object ActualLevel -eq "Danger").Count
                Error = @($items | Where-Object ActualLevel -eq "ERROR").Count
                MinRiskScore = [math]::Round((($items | Where-Object RiskScore -ne "" | ForEach-Object { [double]$_.RiskScore }) | Measure-Object -Minimum).Minimum, 2)
                MaxRiskScore = [math]::Round((($items | Where-Object RiskScore -ne "" | ForEach-Object { [double]$_.RiskScore }) | Measure-Object -Maximum).Maximum, 2)
                AvgRiskScore = [math]::Round((($items | Where-Object RiskScore -ne "" | ForEach-Object { [double]$_.RiskScore }) | Measure-Object -Average).Average, 2)
            }
        } |
        Sort-Object DDL |
        Export-Csv -Path $ddlSummaryPath -NoTypeInformation -Encoding UTF8

    $ddlByConfigPath = Join-Path $runDir "ddl_by_config_summary.csv"
    $results |
        Group-Object DDL, IntendedLevel, Config |
        ForEach-Object {
            $items = $_.Group
            [pscustomobject]@{
                DDL = $items[0].DDL
                IntendedLevel = $items[0].IntendedLevel
                Config = $items[0].Config
                Total = $items.Count
                Safe = @($items | Where-Object ActualLevel -eq "Safe").Count
                Warning = @($items | Where-Object ActualLevel -eq "Warning").Count
                Danger = @($items | Where-Object ActualLevel -eq "Danger").Count
                Error = @($items | Where-Object ActualLevel -eq "ERROR").Count
                MinRiskScore = [math]::Round((($items | Where-Object RiskScore -ne "" | ForEach-Object { [double]$_.RiskScore }) | Measure-Object -Minimum).Minimum, 2)
                MaxRiskScore = [math]::Round((($items | Where-Object RiskScore -ne "" | ForEach-Object { [double]$_.RiskScore }) | Measure-Object -Maximum).Maximum, 2)
                AvgRiskScore = [math]::Round((($items | Where-Object RiskScore -ne "" | ForEach-Object { [double]$_.RiskScore }) | Measure-Object -Average).Average, 2)
            }
        } |
        Sort-Object DDL, Config |
        Export-Csv -Path $ddlByConfigPath -NoTypeInformation -Encoding UTF8

    if (-not $KeepSandboxDbs) {
        Remove-Item -LiteralPath $sandboxDir -Recurse -Force
    }

    Write-Host "[DONE] Matrix results: $matrixPath"
    Write-Host "[DONE] DDL summary:    $ddlSummaryPath"
    Write-Host "[DONE] Config summary: $ddlByConfigPath"
}
finally {
    Pop-Location
}
