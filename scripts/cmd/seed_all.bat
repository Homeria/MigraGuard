@echo off
setlocal enabledelayedexpansion

:: MigraGuard Batch Seeder (Windows CMD)
:: Seeds all scenarios found in experiments\scenarios into experiments\data

set DATA_DIR=experiments\data
set SCENARIO_DIR=experiments\scenarios

if not exist "%DATA_DIR%" mkdir "%DATA_DIR%"

echo 🚀 Starting Batch Seeding for all scenarios...

for %%f in ("%SCENARIO_DIR%\*.yaml") do (
    set SCENARIO_PATH=%%f
    echo --------------------------------------------------
    echo Processing: %%~nxf
    
    go run ./cmd/migraguard simulate --scenario "%%f" --force
    
    :: Extract experiment_name to move the resulting DB
    for /f "tokens=2 delims=: " %%a in ('findstr "experiment_name" "%%f"') do (
        set DB_NAME=%%a
        set DB_NAME=!DB_NAME:"=!
        set DB_NAME=!DB_NAME:'=!
    )
    
    if exist "!DB_NAME!.db" (
        move /y "!DB_NAME!.db" "%DATA_DIR%\"
        echo [OK] Generated: %DATA_DIR%\!DB_NAME!.db
    )
)

echo --------------------------------------------------
echo ✅ Batch seeding complete. All databases are in %DATA_DIR%
pause
