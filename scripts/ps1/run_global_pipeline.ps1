# 🛡️ MigraGuard L3 Orchestrator: Global Multi-Scenario Pipeline (PowerShell Standalone)
# Usage: .\scripts\ps1\run_global_pipeline.ps1

$ModuleDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ModuleDir\.."

Write-Host "=================================================="
Write-Host "🛡️  MigraGuard L3 Global Multi-Scenario Orchestrator"
Write-Host "=================================================="

# 1. Collect all scenarios
$Scenarios = Get-ChildItem -Path (Join-Path $ProjectRoot "experiments\scenarios\*.yaml") -ErrorAction SilentlyContinue
$Total = $Scenarios.Count

if ($Total -eq 0) {
    Write-Error "❌ [ERROR] No scenario files found in experiments/scenarios/"
    exit 1
}

Write-Host "📂 Found $Total scenario file(s) to process."
Write-Host "=================================================="

$Count = 0
$ScenarioScript = Join-Path $ModuleDir "run_scenario_pipeline.ps1"
if (-not (Test-Path $ScenarioScript)) {
    $ScenarioScript = Join-Path $ModuleDir "ps1\run_scenario_pipeline.ps1"
}

foreach ($Scenario in $Scenarios) {
    $Count++
    $SName = $Scenario.Name
    
    Write-Host ""
    Write-Host "=================================================="
    Write-Host "👉 [Scenario $Count/$Total] Processing: $SName"
    Write-Host "=================================================="
    
    # Call L2 Orchestrator for this specific scenario
    & $ScenarioScript $Scenario.FullName
    
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "⚠️  [WARNING] Pipeline execution failed for scenario: $SName. Continuing to next."
    } else {
        Write-Host "✅ [Scenario SUCCESS] Completed scenario: $SName"
    }
}

Write-Host ""
Write-Host "=================================================="
Write-Host "🏆 [L3 SUCCESS] Global Multi-Scenario Pipeline Completed!"
Write-Host "📂 All archived outputs are located in experiments/reports/batch_runs/"
Write-Host "=================================================="
exit 0
