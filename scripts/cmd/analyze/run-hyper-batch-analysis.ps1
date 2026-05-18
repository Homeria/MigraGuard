# 🛡️ MigraGuard: Multi-Dimensional Batch Research Framework
# This script runs all scenarios against all DDL cases with multiple config variations.

$scenariosDir = "experiments/scenarios"
$ddlDir = "experiments/ddl"
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$batchOutputDir = "experiments/reports/batch_runs/$timestamp"
$bin = "./migraguard.exe"
$pythonScript = "tools/visualization/analyze/plot_predictive_heatmap.py"

# --- Configuration Variation Matrix ---
# You can add more parameters here to test different system profiles.
$configVariations = @(
    @{ name = "Default"; mu_max = 5000; disk_io = 104857600 },
    @{ name = "LowCapacity"; mu_max = 2000; disk_io = 52428800 },
    @{ name = "HighCapacity"; mu_max = 10000; disk_io = 209715200 }
)

if (-not (Test-Path $batchOutputDir)) { New-Item -ItemType Directory -Path $batchOutputDir -Force }

$scenarios = Get-ChildItem "$scenariosDir/*.yaml"
$ddls = Get-ChildItem "$ddlDir/*.sql"

Write-Host "🚀 Starting Multi-Dimensional Batch Analysis..." -ForegroundColor Cyan
Write-Host "Scenarios: $($scenarios.Count) | DDLs: $($ddls.Count) | Configs: $($configVariations.Count)"
Write-Host "Total Estimated Runs: $($scenarios.Count * $ddls.Count * $configVariations.Count)"

foreach ($config in $configVariations) {
    $configName = $config.name
    $configDir = Join-Path $batchOutputDir $configName
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
    
    # Save Metadata for this config
    $metadata = @{
        profile_name = $configName
        mu_max = $config.mu_max
        disk_io = $config.disk_io
        timestamp = $timestamp
    }
    $metadata | ConvertTo-Json | Out-File (Join-Path $configDir "config_metadata.json")
    
    # Set Environment Variables for Viper Overrides
    $env:MIGRAGUARD_RISK_MU_MAX = $config.mu_max
    $env:MIGRAGUARD_RISK_DISK_IO = $config.disk_io
    
    Write-Host "`n[Config Profile] $configName (MU_MAX=$($config.mu_max), DISK_IO=$($config.disk_io))" -ForegroundColor Blue

    foreach ($scenario in $scenarios) {
        $scenarioName = [io.path]::GetFileNameWithoutExtension($scenario.Name)
        $scenarioDir = Join-Path $configDir $scenarioName
        New-Item -ItemType Directory -Path $scenarioDir -Force | Out-Null
        
        Write-Host "  [Scenario] $scenarioName" -ForegroundColor Yellow
        
        # 1. Seed Sandbox
        & $bin simulate --scenario $scenario.FullName --force | Out-Null
        
        # Find DB Path
        $experimentName = (Select-String -Path $scenario.FullName -Pattern 'experiment_name:\s*"(.*)"').Matches.Groups[1].Value
        if (-not $experimentName) { $experimentName = (Select-String -Path $scenario.FullName -Pattern 'experiment_name:\s*(.*)').Matches.Groups[1].Value }
        $dbPath = "$experimentName.db"
        
        if (-not (Test-Path $dbPath)) { continue }

        foreach ($ddl in $ddls) {
            $ddlName = [io.path]::GetFileNameWithoutExtension($ddl.Name)
            
            # 2. Analyze
            & $bin analyze $ddl.FullName --sandbox $dbPath --forecast --output console | Out-Null
            
            # 3. Visualize & Move
            if (Test-Path "predictive_forecast.csv") {
                python $pythonScript "predictive_forecast.csv" | Out-Null
                
                if (Test-Path "predictive_risk_heatmap.png") {
                    $destPath = Join-Path $scenarioDir "$ddlName.png"
                    Move-Item -Path "predictive_risk_heatmap.png" -Destination $destPath -Force
                    
                    # Store CSV for future ML training/analysis
                    $csvDestPath = Join-Path $scenarioDir "$ddlName.csv"
                    Move-Item -Path "predictive_forecast.csv" -Destination $csvDestPath -Force
                }
            }
        }
        
        # Cleanup DB
        if (Test-Path $dbPath) { Remove-Item $dbPath -Force }
    }
}

# Clear Environment Variables
$env:MIGRAGUARD_RISK_MU_MAX = $null
$env:MIGRAGUARD_RISK_DISK_IO = $null

Write-Host "`n✅ Multi-dimensional batch analysis complete!" -ForegroundColor Green
Write-Host "Results organized in: $batchOutputDir"
