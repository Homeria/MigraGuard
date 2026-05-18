# 🛡️ MigraGuard: All-in-One Heatmap Batch Generator
# This script runs all scenarios against all DDL cases and generates heatmaps.

$scenariosDir = "experiments/scenarios"
$ddlDir = "experiments/ddl"
$outputBaseDir = "experiments/reports/heatmaps"
$bin = "./migraguard.exe"
$pythonScript = "tools/visualization/analyze/plot_predictive_heatmap.py"

if (-not (Test-Path $outputBaseDir)) { New-Item -ItemType Directory -Path $outputBaseDir }

$scenarios = Get-ChildItem "$scenariosDir/*.yaml"
$ddls = Get-ChildItem "$ddlDir/*.sql"

Write-Host "🚀 Starting Batch Heatmap Generation..." -ForegroundColor Cyan
Write-Host "Found $($scenarios.Count) scenarios and $($ddls.Count) DDL cases. Total: $($scenarios.Count * $ddls.Count) runs."

foreach ($scenario in $scenarios) {
    $scenarioName = [io.path]::GetFileNameWithoutExtension($scenario.Name)
    Write-Host "`n[Scenario] $scenarioName" -ForegroundColor Yellow
    
    # 1. Simulate to ensure DB exists and is fresh
    Write-Host "  -> Seeding sandbox database..."
    & $bin simulate --scenario $scenario.FullName --force | Out-Null
    
    # Extract experiment_name from YAML to find the .db file
    # Simple grep-like approach for YAML
    $experimentName = (Select-String -Path $scenario.FullName -Pattern 'experiment_name:\s*"(.*)"').Matches.Groups[1].Value
    if (-not $experimentName) {
        $experimentName = (Select-String -Path $scenario.FullName -Pattern 'experiment_name:\s*(.*)').Matches.Groups[1].Value
    }
    $dbPath = "$experimentName.db"
    
    if (-not (Test-Path $dbPath)) {
        Write-Host "  [ERROR] Sandbox DB not found: $dbPath" -ForegroundColor Red
        continue
    }

    # Create output directory for this scenario
    $scenarioOutputDir = Join-Path $outputBaseDir $scenarioName
    if (-not (Test-Path $scenarioOutputDir)) { New-Item -ItemType Directory -Path $scenarioOutputDir }

    foreach ($ddl in $ddls) {
        $ddlName = [io.path]::GetFileNameWithoutExtension($ddl.Name)
        Write-Host "    -> Analyzing DDL: $ddlName"
        
        # 2. Analyze & Forecast
        & $bin analyze $ddl.FullName --sandbox $dbPath --forecast --output console | Out-Null
        
        # 3. Visualize
        if (Test-Path "predictive_forecast.csv") {
            python $pythonScript "predictive_forecast.csv" | Out-Null
            
            if (Test-Path "predictive_risk_heatmap.png") {
                $finalName = "$ddlName.png"
                $destPath = Join-Path $scenarioOutputDir $finalName
                Move-Item -Path "predictive_risk_heatmap.png" -Destination $destPath -Force
                # Also clean up the CSV to be safe
                Remove-Item -Path "predictive_forecast.csv" -Force
            } else {
                Write-Host "      [WARNING] Heatmap image was not generated for $ddlName" -ForegroundColor Gray
            }
        } else {
            Write-Host "      [WARNING] Forecast CSV was not generated for $ddlName" -ForegroundColor Gray
        }
    }
    
    # Cleanup DB after scenario is done to save space if needed, 
    # but here we might want to keep them in experiments/data if moved.
    # For now, let's move it to experiments/data if it's in the root
    if (Test-Path $dbPath) {
        Move-Item -Path $dbPath -Destination "experiments/data/$dbPath" -Force
    }
}

Write-Host "`n✅ All heatmaps generated in $outputBaseDir" -ForegroundColor Green
