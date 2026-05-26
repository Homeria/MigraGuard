# 🛡️ MigraGuard L1 Module: Analyze Single DDL (PowerShell Standalone)
# Usage: .\scripts\ps1\modules\analyze_single_ddl.ps1 <ddl_path> <db_path> <config_path> <output_csv_path> [--append]

param (
    [string]$DdlPath,
    [string]$DbPath,
    [string]$ConfigPath,
    [string]$OutCsv,
    [switch]$Append
)

$ModuleDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ModuleDir\..\..\.."

# 1. Guard Clauses
if ([string]::IsNullOrEmpty($DdlPath) -or [string]::IsNullOrEmpty($DbPath) -or [string]::IsNullOrEmpty($ConfigPath) -or [string]::IsNullOrEmpty($OutCsv)) {
    Write-Error "❌ [ERROR] Missing required arguments."
    Write-Host "Usage: .\analyze_single_ddl.ps1 <ddl_path> <db_path> <config_path> <output_csv_path> [-Append]"
    exit 1
}

# Resolve paths to absolute if relative
if (-not [System.IO.Path]::IsPathRooted($DdlPath)) { $DdlPath = Join-Path $ProjectRoot $DdlPath }
if (-not [System.IO.Path]::IsPathRooted($DbPath)) { $DbPath = Join-Path $ProjectRoot $DbPath }
if (-not [System.IO.Path]::IsPathRooted($ConfigPath)) { $ConfigPath = Join-Path $ProjectRoot $ConfigPath }
if (-not [System.IO.Path]::IsPathRooted($OutCsv)) { $OutCsv = Join-Path $ProjectRoot $OutCsv }

if (-not (Test-Path $DdlPath)) { Write-Error "❌ [ERROR] DDL file not found: $DdlPath"; exit 1 }
if (-not (Test-Path $DbPath)) { Write-Error "❌ [ERROR] Database file not found: $DbPath"; exit 1 }
if (-not (Test-Path $ConfigPath)) { Write-Error "❌ [ERROR] Config file not found: $ConfigPath"; exit 1 }

$Mode = "OVERWRITE"
if ($Append) { $Mode = "APPEND" }

# 2. Run analysis
Push-Location $ProjectRoot

$HeaderFlag = ""
if ($Mode -eq "APPEND" -and (Test-Path $OutCsv)) {
    $HeaderFlag = "--no-header"
}

$DestDir = Split-Path $OutCsv -Parent
if (-not (Test-Path $DestDir)) {
    New-Item -ItemType Directory -Force -Path $DestDir | Out-Null
}

$BinaryPath = Join-Path $ProjectRoot "build\migraguard.exe"
if (-not (Test-Path $BinaryPath)) {
    $BinaryPath = Join-Path $ProjectRoot "build\migraguard"
}

# Run execution. Ignore standard error exit code to allow continuous batch flow
if ($Mode -eq "OVERWRITE") {
    & $BinaryPath analyze $DdlPath --sandbox $DbPath --output csv $HeaderFlag --config $ConfigPath | Out-File -FilePath $OutCsv -Encoding utf8
} else {
    & $BinaryPath analyze $DdlPath --sandbox $DbPath --output csv $HeaderFlag --config $ConfigPath | Out-File -FilePath $OutCsv -Encoding utf8 -Append
}

Pop-Location

Write-Host "   -> [L1 SUCCESS] Analyzed $(Split-Path $DdlPath -Leaf) on $(Split-Path $DbPath -Leaf) (Mode: $Mode)"
exit 0
