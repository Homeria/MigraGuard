# 🛡️ MigraGuard L1 Module: Simulate Scenario (PowerShell Standalone)
# Usage: .\scripts\ps1\modules\simulate_scenario.ps1 <scenario_yaml_path>

param (
    [string]$ScenarioPath
)

$ModuleDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ModuleDir\..\..\.."

# 1. Guard Clauses
if ([string]::IsNullOrEmpty($ScenarioPath)) {
    Write-Error "❌ [ERROR] Scenario YAML path is required."
    Write-Host "Usage: .\simulate_scenario.ps1 <scenario_yaml_path>"
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

# 2. Run simulation
Write-Host "🔄 [L1] Simulating Scenario: $(Split-Path $ScenarioPath -Leaf)"
Push-Location $ProjectRoot

# Symmetrical reference: on Windows, we look for build/migraguard.exe
$BinaryPath = Join-Path $ProjectRoot "build\migraguard.exe"
if (-not (Test-Path $BinaryPath)) {
    # fallback to local standard if not compiled with extension
    $BinaryPath = Join-Path $ProjectRoot "build\migraguard"
}

& $BinaryPath simulate --scenario $ScenarioPath --force
$ExitCode = $LASTEXITCODE

Pop-Location

if ($ExitCode -ne 0) {
    Write-Error "❌ [ERROR] Simulation execution failed."
    exit 1
}

Write-Host "✅ [L1 SUCCESS] Simulation complete for $(Split-Path $ScenarioPath -Leaf)"
exit 0
