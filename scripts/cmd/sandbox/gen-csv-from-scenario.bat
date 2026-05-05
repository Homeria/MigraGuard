@echo off
setlocal enabledelayedexpansion

if "%~1"==="" (
    echo Usage: %0 ^<scenario_name^> [output_path] [--force]
    exit /b 1
)

set S_PATH=experiments\scenarios\%~1
if not exist !S_PATH! set S_PATH=%~1

:: Smart Output Detection
set OUT=%~2
set FORCE_FLAG=
if "!OUT!"=="--force" set FORCE_FLAG=--force& set OUT=
if "!OUT!"=="" (
    if not exist experiments\reports\metrics mkdir experiments\reports\metrics
    set OUT=experiments\reports\metrics\%~n1_metrics.csv
)

echo 📊 Generating Metrics CSV from %~1...

go run ./cmd/migraguard simulate --scenario !S_PATH! --csv !OUT! --no-db !FORCE_FLAG!

if !ERRORLEVEL! EQU 0 (
    echo [OK] Metrics CSV successfully generated at: !OUT!
)
