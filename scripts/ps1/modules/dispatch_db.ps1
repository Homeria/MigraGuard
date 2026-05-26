# 🛡️ MigraGuard L1 Module: Dispatch DB to Case (PowerShell Standalone)
# Usage: .\scripts\ps1\modules\dispatch_db.ps1 <seed_db_path> <case_yaml_path> [override_db_path]

param (
    [string]$SeedDbPath,
    [string]$CaseYamlPath,
    [string]$OverrideDbPath
)

$ModuleDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ModuleDir\..\..\.."

# 1. Guard Clauses
if ([string]::IsNullOrEmpty($SeedDbPath) -or [string]::IsNullOrEmpty($CaseYamlPath)) {
    Write-Error "❌ [ERROR] Both seed_db_path and case_yaml_path are required."
    Write-Host "Usage: .\dispatch_db.ps1 <seed_db_path> <case_yaml_path> [override_db_path]"
    exit 1
}

# Resolve to absolute paths if relative
if (-not [System.IO.Path]::IsPathRooted($SeedDbPath)) { $SeedDbPath = Join-Path $ProjectRoot $SeedDbPath }
if (-not [System.IO.Path]::IsPathRooted($CaseYamlPath)) { $CaseYamlPath = Join-Path $ProjectRoot $CaseYamlPath }

if (-not (Test-Path $SeedDbPath)) {
    Write-Error "❌ [ERROR] Seed DB not found: $SeedDbPath"
    exit 1
}
if (-not (Test-Path $CaseYamlPath)) {
    Write-Error "❌ [ERROR] Case YAML configuration not found: $CaseYamlPath"
    exit 1
}

# 2. Extract db_path and case_name dynamically from case configuration metadata block
$YamlContent = Get-Content $CaseYamlPath -Raw

# Match metadata block case_name
$CaseName = ""
if ($YamlContent -match "case_name:\s*['\"" ]?([^'\""\r\n]+)['\"" ]?") {
    $CaseName = $Matches[1].Trim()
}

$DbPath = ""
if (-not [string]::IsNullOrEmpty($OverrideDbPath)) {
    $DbPath = $OverrideDbPath
} else {
    if ($YamlContent -match "db_path:\s*['\"" ]?([^'\""\r\n]+)['\"" ]?") {
        $DbPath = $Matches[1].Trim()
    }
}

if ([string]::IsNullOrEmpty($DbPath)) {
    Write-Error "❌ [ERROR] Could not parse db_path from configuration metadata: $CaseYamlPath"
    exit 1
}

# Resolve destination db_path to absolute path if relative
if (-not [System.IO.Path]::IsPathRooted($DbPath)) {
    $DbPath = Join-Path $ProjectRoot $DbPath
}

# 3. Create destination directory and copy database
$DestDir = Split-Path $DbPath -Parent
if (-not (Test-Path $DestDir)) {
    New-Item -ItemType Directory -Force -Path $DestDir | Out-Null
}

Copy-Item $SeedDbPath $DbPath -Force
if ($LASTEXITCODE -ne 0 -and $?) {
    # Check simple error handling
}

Write-Host "   -> [L1 SUCCESS] Dispatched to Case [$CaseName]: $(Split-Path $DbPath -Leaf)"
exit 0
