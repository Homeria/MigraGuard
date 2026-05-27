# 🛡️ MigraGuard L2 Orchestrator: Single Scenario Pipeline (PowerShell Standalone)
# Usage: .\scripts\ps1\run_scenario_pipeline.ps1 <scenario_yaml_path>

param (
    [string]$ScenarioPath
)

$ModuleDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ModuleDir\.."

# 1. Guard Clauses
if ([string]::IsNullOrEmpty($ScenarioPath)) {
    Write-Error "❌ [ERROR] Scenario YAML path is required."
    Write-Host "Usage: .\run_scenario_pipeline.ps1 <scenario_yaml_path>"
    exit 1
}

# Resolve to absolute path if relative
if (-not [System.IO.Path]::IsPathRooted($ScenarioPath)) {
    $ScenarioPath = Join-Path $ProjectRoot $ScenarioPath
}

if (-not (Test-Path $ScenarioPath)) {
    Write-Error "❌ [ERROR] Scenario file not found: $ScenarioPath"
    exit 1
}

Write-Host "=================================================="
Write-Host "🛡️  MigraGuard L2 Scenario Pipeline: $(Split-Path $ScenarioPath -Leaf)"
Write-Host "=================================================="

# 0. Define Unique runs package directory
$Timestamp = Get-Date -Format "yyyy-MM-dd_HH-mm-ss"
$RunDir = Join-Path $ProjectRoot "experiments\reports\batch_runs\$Timestamp"
$RunDataDir = Join-Path $RunDir "data"
if (-not (Test-Path $RunDataDir)) {
    New-Item -ItemType Directory -Force -Path $RunDataDir | Out-Null
}

# 0. Clean old py scripts
Write-Host "🧹 [0/5] Cleaning temporary Python files inside reports..."
Get-ChildItem -Path (Join-Path $ProjectRoot "experiments\reports") -Filter "*.py" -Recurse -ErrorAction SilentlyContinue | Remove-Item -Force -ErrorAction SilentlyContinue

# 1. Simulate seed
$SimulateScript = Join-Path $ModuleDir "ps1\modules\simulate_scenario.ps1"
if (-not (Test-Path $SimulateScript)) {
    $SimulateScript = Join-Path $ModuleDir "modules\simulate_scenario.ps1"
}
& $SimulateScript $ScenarioPath
if ($LASTEXITCODE -ne 0) {
    Write-Error "❌ [ERROR] Simulation stage failed."
    exit 1
}

# Extract seed DB name
$YamlContent = Get-Content $ScenarioPath -Raw
$ExpName = ""
if ($YamlContent -match "experiment_name:\s*['\"" ]?([^'\""\r\n]+)['\"" ]?") {
    $ExpName = $Matches[1].Trim()
}

$SeedDb = Join-Path $ProjectRoot "$ExpName.db"
if (-not (Test-Path $SeedDb)) {
    $SeedDb = Join-Path $ProjectRoot "experiments\data\$ExpName.db"
}

if (-not (Test-Path $SeedDb)) {
    Write-Error "❌ [ERROR] Seed database not found: $SeedDb"
    exit 1
}

# 2. Dispatch DB to each case nested inside the runs pack (using override parameter)
Write-Host "📂 [2/5] Dispatching databases to capacity cases nested in runs pack..."
$Configs = Get-ChildItem -Path (Join-Path $ProjectRoot "experiments\configs\cases\*.yaml") -ErrorAction SilentlyContinue

$DispatchScript = Join-Path $ModuleDir "ps1\modules\dispatch_db.ps1"
if (-not (Test-Path $DispatchScript)) {
    $DispatchScript = Join-Path $ModuleDir "modules\dispatch_db.ps1"
}

foreach ($Config in $Configs) {
    $ConfContent = Get-Content $Config.FullName -Raw
    $CaseName = ""
    if ($ConfContent -match "case_name:\s*['\"" ]?([^'\""\r\n]+)['\"" ]?") {
        $CaseName = $Matches[1].Trim()
    }
    $DestDb = Join-Path $RunDataDir "$CaseName.db"
    & $DispatchScript $SeedDb $Config.FullName $DestDb
}

# Cleanup root seed db copy if it exists to keep workspace tidy
if (Test-Path (Join-Path $ProjectRoot "$ExpName.db")) {
    Remove-Item (Join-Path $ProjectRoot "$ExpName.db") -Force -ErrorAction SilentlyContinue
}

# 3. Batch Analyze (All Case configs x All DDL sqls) targeting runs DBs and local runs CSV inside data/
Write-Host "🚀 [3/5] Executing dynamic batch analyses..."
$ReportCsv = Join-Path $RunDataDir "research_results.csv"
if (Test-Path $ReportCsv) { Remove-Item $ReportCsv -Force }

$Sqls = Get-ChildItem -Path (Join-Path $ProjectRoot "experiments\ddl\*.sql") -ErrorAction SilentlyContinue
$Count = 0

$AnalyzeScript = Join-Path $ModuleDir "ps1\modules\analyze_single_ddl.ps1"
if (-not (Test-Path $AnalyzeScript)) {
    $AnalyzeScript = Join-Path $ModuleDir "modules\analyze_single_ddl.ps1"
}

foreach ($Config in $Configs) {
    $ConfContent = Get-Content $Config.FullName -Raw
    $CaseName = ""
    if ($ConfContent -match "case_name:\s*['\"" ]?([^'\""\r\n]+)['\"" ]?") {
        $CaseName = $Matches[1].Trim()
    }
    $DbPath = Join-Path $RunDataDir "$CaseName.db"

    foreach ($Sql in $Sqls) {
        $Count++
        $AppendFlag = $false
        if ($Count -gt 1) { $AppendFlag = $true }

        if ($AppendFlag) {
            & $AnalyzeScript $Sql.FullName $DbPath $Config.FullName $ReportCsv -Append
        } else {
            & $AnalyzeScript $Sql.FullName $DbPath $Config.FullName $ReportCsv
        }
    }
}

# 4. Generate heatmaps directly inside the runs pack
Write-Host "📊 [4/5] Plotting 24h Individual heatmaps directly inside runs pack..."
python3 (Join-Path $ProjectRoot "tools\visualization\analyze\run_massive_forecast_plots.py") $RunDir

# 5. Plot Master report directly inside the data/ folder of runs pack
Write-Host "🎨 [5/5] Generating Master Heatmap Report directly inside data/ folder..."
python3 (Join-Path $ProjectRoot "tools\visualization\analyze\plot_research_report.py") $ReportCsv

# Sync results to artifact path for system integration
$ArtifactSyncPath = "/home/gyeongho/.gemini/antigravity-cli/brain/fc61e4a8-7c53-4959-8257-7d147633cdc3/"
if (Test-Path $ArtifactSyncPath) {
    Copy-Item (Join-Path $RunDataDir "research_results_analysis.png") $ArtifactSyncPath -Force -ErrorAction SilentlyContinue
}

Write-Host "=================================================="
Write-Host "✅ [L2 SUCCESS] Scenario Pipeline Complete!"
Write-Host "📍 Master Chart: $(Join-Path $RunDataDir "research_results_analysis.png")"
Write-Host "=================================================="
exit 0
